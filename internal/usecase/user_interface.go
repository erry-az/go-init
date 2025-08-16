package usecase

import (
	"context"

	"github.com/erry-az/go-sample/internal/domain"
)

// UserUsecase defines the business logic interface for user operations
type UserUsecase interface {
	CreateUser(ctx context.Context, name, email string) (*domain.User, error)
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	UpdateUser(ctx context.Context, userID, name, email string) (*domain.User, error)
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	BulkCreateUsers(ctx context.Context, users []BulkCreateUserRequest) (*BulkCreateUsersResponse, error)
}

// Request/Response types for operations that need multiple parameters
type ListUsersRequest struct {
	PageSize    int32
	PageToken   string
	SearchQuery string
}

type ListUsersResponse struct {
	Users         []*domain.User
	NextPageToken string
	TotalCount    int32
}

type BulkCreateUserRequest struct {
	Name  string
	Email string
}

type BulkCreateUsersResponse struct {
	Users        []*domain.User
	FailedEmails []string
}
