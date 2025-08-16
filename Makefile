.PHONY: all build clean test lint generate proto sqlc mocks migrate new-migration migration-status up down restart stop reset run dev check setup status menu help

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

## Run database migrations using Docker
migrate:
	@echo "🔄 Running database migrations..."
	docker compose run --rm migrate migrate apply --env local

## Create new migration file using Docker
new-migration:
	@echo "➕ Creating new migration..."
	docker compose run --rm migrate migrate hash --env local
	@read -p "Enter migration name: " name; \
	docker compose run --rm migrate migrate diff $$name --env local

## Check migration status using Docker
migration-status:
	@echo "📊 Checking migration status..."
	docker compose run --rm migrate migrate status --env local

## Start development environment
dev:
	@echo "🐳 Starting development environment..."
	docker compose up --build --watch && docker image prune -f

## Start services in background
up:
	@echo "🐳 Starting services in background..."
	docker compose up --build -d && docker image prune -f

## Stop services
stop:
	@echo "🛑 Stopping services..."
	docker compose down

## Restart services
restart:
	@echo "🔄 Restarting services..."
	docker compose down && docker compose up --build -d && docker image prune -f

## Reset all data (stop + remove volumes)
reset:
	@echo "🗑️ Resetting all data..."
	docker compose down -v

## Run application locally (without Docker)
run:
	@echo "🚀 Running application locally..."
	go run ./cmd/server

## Quick code validation
check: lint test

## Full project setup
setup: generate build test

## Show project status (git + docker services)
status:
	@echo "📋 Project Status"
	@echo "=================="
	@echo "Git Status:"
	@git status --short
	@echo ""
	@echo "Docker Services:"
	@docker-compose ps 2>/dev/null || echo "Docker services not running"

## Launch interactive development menu (alias for dev)
menu:
	@./tools/scripts/dev-menu.sh

## Initialize this repo as a template for new projects
template-init:
	@go run ./cmd/template-init
	go mod tidy
	make generate
	make test

## Show all available Makefile targets with descriptions
help:
	@echo "🚀 Available Commands:"
	@echo "====================="
	@awk '/^##/ { desc = substr($$0, 4); getline; if ($$0 ~ /^[a-zA-Z0-9_-]+:/) { target = $$1; gsub(/:.*/, "", target); printf "  %-20s %s\n", target, desc } }' $(MAKEFILE_LIST)