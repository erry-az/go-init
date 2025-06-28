package repository

import (
	"context"

	"github.com/erry-az/go-init/db/sqlc"
	"github.com/google/uuid"
)

// PostgresRepository implements the Repository interface using PostgreSQL
type PostgresRepository struct {
	queries *sqlc.Queries
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(queries *sqlc.Queries) *PostgresRepository {
	return &PostgresRepository{
		queries: queries,
	}
}

// CreateUser creates a new user in the database
func (r *PostgresRepository) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
	return r.queries.CreateUser(ctx, params)
}

// GetUser retrieves a user by ID
func (r *PostgresRepository) GetUser(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUser(ctx, id)
}

// ListUsers retrieves all users
func (r *PostgresRepository) ListUsers(ctx context.Context) ([]sqlc.User, error) {
	return r.queries.ListUsers(ctx)
}

// UpdateUser updates a user
func (r *PostgresRepository) UpdateUser(ctx context.Context, params sqlc.UpdateUserParams) (sqlc.User, error) {
	return r.queries.UpdateUser(ctx, params)
}

// DeleteUser deletes a user
func (r *PostgresRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, id)
}
