package repository

import (
	"context"

	"github.com/erry-az/go-init/db/sqlc"
	"github.com/google/uuid"
)

// Repository defines the interface for database operations
type Repository interface {
	CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	ListUsers(ctx context.Context) ([]sqlc.User, error)
	UpdateUser(ctx context.Context, params sqlc.UpdateUserParams) (sqlc.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
