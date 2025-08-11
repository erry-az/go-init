package grpc

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"

	sqlc2 "github.com/erry-az/go-init/internal/repository/sqlc"
	"github.com/erry-az/go-init/pkg/watermill"
	"github.com/erry-az/go-init/proto/api/v1"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	v1.UnimplementedUserServiceServer
	db        sqlc2.Querier
	publisher *watermill.Publisher
}

func NewUserService(db sqlc2.Querier, publisher *watermill.Publisher) *UserService {
	return &UserService{
		db:        db,
		publisher: publisher,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	userID := uuid.New()

	params := sqlc2.CreateUserParams{
		ID:    userID,
		Name:  req.Name,
		Email: req.Email,
	}

	dbUser, err := s.db.CreateUser(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", req.Email)
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	user := s.mapDBUserToProto(dbUser)

	// Publish user created event
	event := &eventv1.UserCreatedEvent{
		EventId:       uuid.New().String(),
		User:          user,
		EventTime:     timestamppb.Now(),
		CorrelationId: s.getCorrelationID(ctx),
		Data: &eventv1.UserCreatedEventData{
			Source: "user-service",
			Metadata: map[string]string{
				"operation": "create_user",
				"version":   "v1",
			},
		},
	}

	if err := s.publisher.PublishProtoMessage(ctx, watermill.TopicUserCreated, event); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish user created event: %v\n", err)
	}

	return &v1.CreateUserResponse{User: user}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	dbUser, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	user := s.mapDBUserToProto(dbUser)
	return &v1.GetUserResponse{User: user}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Check if user exists
	_, err = s.db.GetUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to check user existence: %v", err)
	}

	params := sqlc2.UpdateUserParams{
		ID:    userID,
		Name:  req.Name,
		Email: req.Email,
	}

	dbUser, err := s.db.UpdateUser(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", req.Email)
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	user := s.mapDBUserToProto(dbUser)

	// Publish user updated event
	event := &eventv1.UserUpdatedEvent{
		EventId:       uuid.New().String(),
		User:          user,
		EventTime:     timestamppb.Now(),
		CorrelationId: s.getCorrelationID(ctx),
		Data: &eventv1.UserUpdatedEventData{
			Source:        "user-service",
			ChangedFields: []string{"name", "email"}, // Could be more dynamic
			Metadata: map[string]string{
				"operation": "update_user",
				"version":   "v1",
			},
		},
	}

	if err := s.publisher.PublishProtoMessage(ctx, watermill.TopicUserUpdated, event); err != nil {
		fmt.Printf("Failed to publish user updated event: %v\n", err)
	}

	return &v1.UpdateUserResponse{User: user}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	// Get user before deletion for event
	dbUser, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	err = s.db.DeleteUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	user := s.mapDBUserToProto(dbUser)

	// Publish user deleted event
	event := &eventv1.UserDeletedEvent{
		EventId:       uuid.New().String(),
		User:          user,
		EventTime:     timestamppb.Now(),
		CorrelationId: s.getCorrelationID(ctx),
		Data: &eventv1.UserDeletedEventData{
			Source: "user-service",
			Reason: "manual_deletion",
			Metadata: map[string]string{
				"operation": "delete_user",
				"version":   "v1",
			},
		},
	}

	if err := s.publisher.PublishProtoMessage(ctx, watermill.TopicUserDeleted, event); err != nil {
		fmt.Printf("Failed to publish user deleted event: %v\n", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := int32(0)
	if req.PageToken != "" {
		decodedOffset, err := base64.StdEncoding.DecodeString(req.PageToken)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token")
		}
		if _, err := fmt.Sscanf(string(decodedOffset), "%d", &offset); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token format")
		}
	}

	var dbUsers []sqlc2.User
	var err error

	if req.SearchQuery != "" {
		params := sqlc2.SearchUsersParams{
			Limit:       pageSize + 1, // Get one extra to check if there are more pages
			Offset:      offset,
			SearchQuery: "%" + req.SearchQuery + "%",
		}
		dbUsers, err = s.db.SearchUsers(ctx, params)
	} else {
		params := sqlc2.ListUsersParams{
			Limit:  pageSize + 1,
			Offset: offset,
		}
		dbUsers, err = s.db.ListUsers(ctx, params)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	// Check if there are more pages
	hasNextPage := len(dbUsers) > int(pageSize)
	if hasNextPage {
		dbUsers = dbUsers[:pageSize] // Remove the extra record
	}

	users := make([]*v1.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = s.mapDBUserToProto(dbUser)
	}

	var nextPageToken string
	if hasNextPage {
		nextOffset := offset + pageSize
		nextPageToken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", nextOffset)))
	}

	// Get total count for response
	var totalCount int32
	if req.SearchQuery != "" {
		count, err := s.db.CountUsersBySearch(ctx, "%"+req.SearchQuery+"%")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to count users: %v", err)
		}
		totalCount = int32(count)
	} else {
		count, err := s.db.CountUsers(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to count users: %v", err)
		}
		totalCount = int32(count)
	}

	return &v1.ListUsersResponse{
		Users:         users,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

func (s *UserService) BulkCreateUsers(ctx context.Context, req *v1.BulkCreateUsersRequest) (*v1.BulkCreateUsersResponse, error) {
	var users []*v1.User
	var failedEmails []string

	for _, userReq := range req.Users {
		createResp, err := s.CreateUser(ctx, userReq)
		if err != nil {
			failedEmails = append(failedEmails, userReq.Email)
			continue
		}
		users = append(users, createResp.User)
	}

	return &v1.BulkCreateUsersResponse{
		Users:        users,
		FailedEmails: failedEmails,
	}, nil
}

func (s *UserService) mapDBUserToProto(dbUser sqlc2.User) *v1.User {
	return &v1.User{
		Id:        dbUser.ID.String(),
		Name:      dbUser.Name,
		Email:     dbUser.Email,
		CreatedAt: timestamppb.New(dbUser.CreatedAt.Time),
		UpdatedAt: timestamppb.New(dbUser.UpdatedAt.Time),
	}
}

func (s *UserService) getCorrelationID(ctx context.Context) string {
	// Try to get correlation ID from context metadata
	// This is a placeholder - in a real app you'd extract this from gRPC metadata
	return uuid.New().String()
}
