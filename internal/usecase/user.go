package usecase

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/erry-az/go-init/internal/domain"
	"github.com/erry-az/go-init/internal/repository/sqlc"
	"github.com/erry-az/go-init/proto/api/v1"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userUsecase struct {
	db        sqlc.Querier
	publisher *cqrs.EventBus
}

// NewUserUsecase creates a new user usecase instance
func NewUserUsecase(db sqlc.Querier, publisher *cqrs.EventBus) UserUsecase {
	return &userUsecase{
		db:        db,
		publisher: publisher,
	}
}

func (u *userUsecase) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	// Create domain entity
	user := domain.NewUser(name, email)

	// Convert to database params
	params := sqlc.CreateUserParams{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	dbUser, err := u.db.CreateUser(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, domain.NewConflictError(fmt.Sprintf("user with email %s already exists", email))
		}
		return nil, domain.NewInternalError(fmt.Sprintf("failed to create user: %v", err))
	}

	// Convert back to domain entity
	createdUser := u.mapDBUserToDomain(dbUser)

	// Publish user created event
	if err := u.publishUserCreatedEvent(ctx, createdUser); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish user created event: %v\n", err)
	}

	return createdUser, nil
}

func (u *userUsecase) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.NewValidationError(fmt.Sprintf("invalid user ID: %v", err))
	}

	dbUser, err := u.db.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewNotFoundError("user not found")
		}
		return nil, domain.NewInternalError(fmt.Sprintf("failed to get user: %v", err))
	}

	return u.mapDBUserToDomain(dbUser), nil
}

func (u *userUsecase) UpdateUser(ctx context.Context, userID, name, email string) (*domain.User, error) {
	// Get existing user
	user, err := u.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update domain entity
	user.UpdateDetails(name, email)

	// Convert to database params
	params := sqlc.UpdateUserParams{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	dbUser, err := u.db.UpdateUser(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, domain.NewConflictError(fmt.Sprintf("user with email %s already exists", email))
		}
		return nil, domain.NewInternalError(fmt.Sprintf("failed to update user: %v", err))
	}

	updatedUser := u.mapDBUserToDomain(dbUser)

	// Publish user updated event
	if err := u.publishUserUpdatedEvent(ctx, updatedUser); err != nil {
		fmt.Printf("Failed to publish user updated event: %v\n", err)
	}

	return updatedUser, nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, userID string) error {
	// Get user before deletion for event
	user, err := u.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	if err := u.db.DeleteUser(ctx, user.ID); err != nil {
		return domain.NewInternalError(fmt.Sprintf("failed to delete user: %v", err))
	}

	// Publish user deleted event
	if err := u.publishUserDeletedEvent(ctx, user); err != nil {
		fmt.Printf("Failed to publish user deleted event: %v\n", err)
	}

	return nil
}

func (u *userUsecase) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
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
			return nil, domain.NewValidationError("invalid page token")
		}
		if _, err := fmt.Sscanf(string(decodedOffset), "%d", &offset); err != nil {
			return nil, domain.NewValidationError("invalid page token format")
		}
	}

	var dbUsers []sqlc.User
	var err error

	if req.SearchQuery != "" {
		params := sqlc.SearchUsersParams{
			Limit:       pageSize + 1,
			Offset:      offset,
			SearchQuery: "%" + req.SearchQuery + "%",
		}
		dbUsers, err = u.db.SearchUsers(ctx, params)
	} else {
		params := sqlc.ListUsersParams{
			Limit:  pageSize + 1,
			Offset: offset,
		}
		dbUsers, err = u.db.ListUsers(ctx, params)
	}

	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to list users: %v", err))
	}

	// Check if there are more pages
	hasNextPage := len(dbUsers) > int(pageSize)
	if hasNextPage {
		dbUsers = dbUsers[:pageSize]
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = u.mapDBUserToDomain(dbUser)
	}

	var nextPageToken string
	if hasNextPage {
		nextOffset := offset + pageSize
		nextPageToken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", nextOffset)))
	}

	// Get total count
	var totalCount int32
	if req.SearchQuery != "" {
		count, err := u.db.CountUsersBySearch(ctx, "%"+req.SearchQuery+"%")
		if err != nil {
			return nil, domain.NewInternalError(fmt.Sprintf("failed to count users: %v", err))
		}
		totalCount = int32(count)
	} else {
		count, err := u.db.CountUsers(ctx)
		if err != nil {
			return nil, domain.NewInternalError(fmt.Sprintf("failed to count users: %v", err))
		}
		totalCount = int32(count)
	}

	return &ListUsersResponse{
		Users:         users,
		NextPageToken: nextPageToken,
		TotalCount:    totalCount,
	}, nil
}

func (u *userUsecase) BulkCreateUsers(ctx context.Context, users []BulkCreateUserRequest) (*BulkCreateUsersResponse, error) {
	var createdUsers []*domain.User
	var failedEmails []string

	for _, userReq := range users {
		user, err := u.CreateUser(ctx, userReq.Name, userReq.Email)
		if err != nil {
			failedEmails = append(failedEmails, userReq.Email)
			continue
		}
		createdUsers = append(createdUsers, user)
	}

	return &BulkCreateUsersResponse{
		Users:        createdUsers,
		FailedEmails: failedEmails,
	}, nil
}

// Helper methods
func (u *userUsecase) mapDBUserToDomain(dbUser sqlc.User) *domain.User {
	return &domain.User{
		ID:        dbUser.ID,
		Name:      dbUser.Name,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}
}

func (u *userUsecase) publishUserCreatedEvent(ctx context.Context, user *domain.User) error {
	event := &eventv1.UserCreatedEvent{
		EventId:       uuid.New().String(),
		User:          u.domainUserToProto(user),
		EventTime:     timestamppb.Now(),
		CorrelationId: u.getCorrelationID(ctx),
		Data: &eventv1.UserCreatedEventData{
			Source: "user-service",
			Metadata: map[string]string{
				"operation": "create_user",
				"version":   "v1",
			},
		},
	}
	return u.publisher.Publish(ctx, event)
}

func (u *userUsecase) publishUserUpdatedEvent(ctx context.Context, user *domain.User) error {
	event := &eventv1.UserUpdatedEvent{
		EventId:       uuid.New().String(),
		User:          u.domainUserToProto(user),
		EventTime:     timestamppb.Now(),
		CorrelationId: u.getCorrelationID(ctx),
		Data: &eventv1.UserUpdatedEventData{
			Source:        "user-service",
			ChangedFields: []string{"name", "email"},
			Metadata: map[string]string{
				"operation": "update_user",
				"version":   "v1",
			},
		},
	}
	return u.publisher.Publish(ctx, event)
}

func (u *userUsecase) publishUserDeletedEvent(ctx context.Context, user *domain.User) error {
	event := &eventv1.UserDeletedEvent{
		EventId:       uuid.New().String(),
		User:          u.domainUserToProto(user),
		EventTime:     timestamppb.Now(),
		CorrelationId: u.getCorrelationID(ctx),
		Data: &eventv1.UserDeletedEventData{
			Source: "user-service",
			Reason: "manual_deletion",
			Metadata: map[string]string{
				"operation": "delete_user",
				"version":   "v1",
			},
		},
	}
	return u.publisher.Publish(ctx, event)
}

func (u *userUsecase) domainUserToProto(user *domain.User) *v1.User {
	return &v1.User{
		Id:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func (u *userUsecase) getCorrelationID(ctx context.Context) string {
	// Try to get correlation ID from context metadata
	// This is a placeholder - in a real app you'd extract this from gRPC metadata
	return uuid.New().String()
}