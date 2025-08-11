.PHONY: all build clean test lint generate proto sqlc mocks db-migrate db-migration-create db-migrate-validate db-migrate-status db-schema-inspect up down run status dev menu help

## Default target - generate code and build application
all: generate build

## Initialize project - setup environment and generate code
init:
	@echo "🚀 Initializing project..."
	@nix-shell --run "make generate"

## Build the application binary
build:
	@echo "🏗️ Building application..."
	mkdir -p bin
	go build -o bin/server ./cmd/server

## Clean build artifacts and generated code  
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin
	rm -f api/proto/*.pb.go api/proto/*.pb.gw.go api/swagger/*.swagger.json
	rm -f db/sqlc/*.sql.go
	rm -rf mocks/*

## Run all tests with verbose output
test:
	@echo "🧪 Running tests..."
	go test -v ./...

## Run golangci-lint on the codebase
lint:
	@echo "✨ Running linter..."
	golangci-lint run ./...

## Generate all code (protobuf, sqlc, mocks)
generate: proto sqlc mocks

## Generate Go code from protobuf definitions
proto:
	@echo "📦 Generating protobuf code..."
	buf generate

## Generate type-safe SQL code using sqlc
sqlc:
	@echo "🗄️ Generating SQL code..."
	sqlc generate

## Generate mocks for testing using mockgen
mocks:
	@echo "🎭 Generating mocks..."
	@if [ -f internal/repository/repository.go ]; then \
		~/go/bin/mockgen -source=internal/repository/repository.go -destination=mocks/mock_repository.go -package=mocks; \
	else \
		echo "Repository interface not found, skipping mock generation"; \
	fi

## Apply pending database migrations
db-migrate:
	@echo "🔄 Running database migrations..."
	atlas migrate apply --env local

## Create a new database migration file
db-migration-create:
	@echo "➕ Creating new migration..."
	atlas migrate hash --env local
	@read -p "Enter migration name: " name; \
	atlas migrate diff $$name --env local

## Validate database migration files
db-migrate-validate:
	@echo "✅ Validating migrations..."
	atlas migrate validate --env local

## Check database migration status
db-migrate-status:
	@echo "📊 Checking migration status..."
	atlas migrate status --env local

## Inspect database and generate schema file
db-schema-inspect:
	@echo "🔍 Inspecting database schema..."
	atlas schema inspect --env local > db/schema.sql

## Start PostgreSQL and RabbitMQ services
up:
	@echo "🐳 Starting services..."
	docker-compose up -d

## Stop all Docker services
down:
	@echo "🛑 Stopping services..."
	docker-compose down

## Start services and run the application
run: up
	@echo "🚀 Running application..."
	go run ./cmd/server

## Show project status (git + docker services)
status:
	@echo "📋 Project Status"
	@echo "=================="
	@echo "Git Status:"
	@git status --short
	@echo ""
	@echo "Docker Services:"
	@docker-compose ps 2>/dev/null || echo "Docker services not running"

## Launch interactive development menu with fzf
dev:
	@./tools/scripts/dev-menu.sh

## Launch interactive development menu (alias for dev)
menu: dev

## Show all available Makefile targets with descriptions
help:
	@echo "🚀 Available Commands:"
	@echo "====================="
	@awk '/^##/ { desc = substr($$0, 4); getline; if ($$0 ~ /^[a-zA-Z0-9_-]+:/) { target = $$1; gsub(/:.*/, "", target); printf "  %-20s %s\n", target, desc } }' $(MAKEFILE_LIST)