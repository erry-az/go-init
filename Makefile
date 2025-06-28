.PHONY: all build clean test lint generate proto sqlc mocks db-migrate db-migration-create up down run

# Default goal
all: generate build

# Build the application
build:
	mkdir -p bin
	go build -o bin/server ./cmd/server

# Clean build artifacts
clean:
	rm -rf bin
	rm -f api/proto/*.pb.go api/proto/*.pb.gw.go api/swagger/*.swagger.json
	rm -f db/sqlc/*.sql.go
	rm -rf mocks/*

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run ./...

# Generate all code
generate: proto sqlc mocks

# Generate code from protobuf definitions
proto:
	buf generate

# Generate SQL code using sqlc
sqlc:
	sqlc generate

# Generate mocks for testing
mocks:
	mockgen -source=internal/repository/repository.go -destination=mocks/mock_repository.go -package=mocks

# Run database migrations
db-migrate:
	goose -dir db/migrations postgres "postgres://postgres:postgres@localhost:5432/go_init_db?sslmode=disable" up

# Create a new migration file
db-migration-create:
	@read -p "Enter migration name: " name; \
	goose -dir db/migrations create $$name sql

# Start services with Docker Compose
up:
	docker-compose up -d

# Stop services with Docker Compose
down:
	docker-compose down

# Run the application in development mode
run:
	go run ./cmd/server
