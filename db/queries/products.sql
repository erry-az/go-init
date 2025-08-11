-- name: CreateProduct :one
INSERT INTO products (
    id,
    name,
    price
) VALUES (
    @id,
    @name,
    @price
) RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = @id;

-- name: ListProducts :many
SELECT * FROM products
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: SearchProducts :many
SELECT * FROM products
WHERE name ILIKE @search_query
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: ListProductsByPriceRange :many
SELECT * FROM products
WHERE price BETWEEN @min_price AND @max_price
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: SearchProductsWithPriceRange :many
SELECT * FROM products
WHERE name ILIKE @search_query AND price BETWEEN @min_price AND @max_price
ORDER BY created_at
LIMIT $1 OFFSET $2;

-- name: CountProducts :one
SELECT COUNT(*) FROM products;

-- name: CountProductsBySearch :one
SELECT COUNT(*) FROM products
WHERE name ILIKE @search_query;

-- name: GetAveragePrice :one
SELECT COALESCE(AVG(price), 0) FROM products;

-- name: GetMinPrice :one
SELECT COALESCE(MIN(price), 0) FROM products;

-- name: GetMaxPrice :one
SELECT COALESCE(MAX(price), 0) FROM products;

-- name: UpdateProduct :one
UPDATE products
SET 
    name = @name,
    price = @price,
    updated_at = NOW()
WHERE id = @id
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = @id;