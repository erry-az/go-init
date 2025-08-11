-- name: CreateUser :one
INSERT INTO users (
    id,
    name,
    email
) VALUES (
    @id,
    @name,
    @email
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = @id;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: SearchUsers :many
SELECT * FROM users
WHERE name ILIKE @search_query OR email ILIKE @search_query
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountUsersBySearch :one
SELECT COUNT(*) FROM users
WHERE name ILIKE @search_query OR email ILIKE @search_query;

-- name: UpdateUser :one
UPDATE users
SET 
    name = @name,
    email = @email,
    updated_at = NOW()
WHERE id = @id
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = @id;