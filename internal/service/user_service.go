package service

import (
	"buf.build/go/protovalidate"
	"context"
	"time"

	apiv1 "github.com/erry-az/go-init/api/v1"
	"github.com/erry-az/go-init/db/sqlc"
	"github.com/erry-az/go-init/internal/repository"
	"github.com/erry-az/go-init/pkg/rabbitmq"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UserService implements the gRPC UserService
type UserService struct {
	apiv1.UnimplementedUserServiceServer
	repo      repository.Repository
	publisher *rabbitmq.Publisher
}

// NewUserService creates a new UserService
func NewUserService(repo repository.Repository, publisher *rabbitmq.Publisher) *UserService {
	return &UserService{
		repo:      repo,
		publisher: publisher,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *apiv1.CreateUserRequest) (*apiv1.User, error) {
	err := protovalidate.Validate(req)
	if err != nil {
		return nil, err
	}

	params := sqlc.CreateUserParams{
		Name:  req.Name,
		Email: req.Email,
	}

	user, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Publish user created event
	event := &apiv1.UserCreatedEvent{
		Id:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
	}
	if err := s.publisher.PublishUserCreated(ctx, event); err != nil {
		// Log error but continue - event publishing should not affect API response
		// In production, you might use a retry mechanism or queue
		// log.Printf("failed to publish user created event: %v", err)
	}

	return &apiv1.User{
		Id:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, req *apiv1.GetUserRequest) (*apiv1.User, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &apiv1.User{
		Id:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Time.Format(time.RFC3339),
	}, nil
}

// ListUsers retrieves all users
func (s *UserService) ListUsers(ctx context.Context, _ *emptypb.Empty) (*apiv1.ListUsersResponse, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	response := &apiv1.ListUsersResponse{
		Users: make([]*apiv1.User, len(users)),
	}

	for i, user := range users {
		response.Users[i] = &apiv1.User{
			Id:        user.ID.String(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Time.Format(time.RFC3339),
		}
	}

	return response, nil
}
