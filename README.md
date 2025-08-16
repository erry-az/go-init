# Go Microservice Template

A modern Go microservice template with complete CRUD operations for users and products, featuring event-driven architecture and clean architecture patterns.

## Features

- Go modules with dependency management
- gRPC + gRPC-Gateway for HTTP/JSON and gRPC APIs
- OpenAPI/Swagger documentation generated from protobuf
- PostgreSQL database with sqlc for type-safe SQL
- Watermill for event-driven messaging (PostgreSQL-based message queue)
- Protocol Buffer validation using buf.build's protovalidate
- Event generation using voi-oss/protoc-gen-event
- Docker Compose for local development with live reload
- Nix shell for development environment
- Makefile for common tasks
- Database migrations with Atlas
- Clean architecture with domain/usecase/handler layers

## Requirements

- [Nix](https://nixos.org/download.html)
- Docker and Docker Compose

## Development

```shell
# Enter the development environment with Nix
nix-shell

# Launch interactive development menu (recommended)
./tools/scripts/dev-menu.sh

# Or run individual commands:
# Start development environment with live reload
make dev

# Start services in background (PostgreSQL databases)
make up

# Run database migrations
make migrate

# Generate code from proto files
make generate

# Run the service locally (without Docker)
make run

# Initialize as template for new projects
make template-init
```

## Project Structure

```
.
├── cmd/                # Application entry points
│   ├── server/         # Main HTTP+gRPC server
│   ├── consumer/       # Event consumer service
│   └── template-init/  # Template initialization tool
├── config/             # Configuration management
├── db/                 # Database related code
│   ├── migrations/     # Atlas database migrations
│   ├── queries/        # SQL queries for sqlc
│   └── schema.sql      # Database schema definition
├── docs/               # Generated OpenAPI/Swagger documentation
│   ├── api/v1/         # API documentation
│   └── event/v1/       # Event documentation
├── internal/           # Private application code
│   ├── domain/         # Domain entities and errors
│   ├── usecase/        # Business logic layer
│   ├── handler/        # HTTP/gRPC/Event handlers
│   ├── repository/     # Data access layer (sqlc generated)
│   ├── server/         # Server implementations
│   └── app/            # Application assembly
├── pkg/                # Public libraries
│   └── watmil/         # Watermill message queue integration
├── proto/              # Protocol Buffer definitions
│   ├── api/v1/         # gRPC service definitions
│   ├── event/v1/       # Event definitions
│   └── options/        # Protobuf options
└── tools/              # Development tools and scripts
    ├── scripts/        # Shell scripts for development
    └── config/nix/     # Nix configuration files
```

## Architecture

This template implements Clean Architecture principles with the following layers:

### Domain Layer (`internal/domain/`)
- Contains business entities (`User`, `Product`) and domain errors
- Pure Go structs with no external dependencies

### Use Case Layer (`internal/usecase/`)  
- Contains business logic and interfaces
- Implements CRUD operations for users and products
- Publishes domain events using Watermill

### Handler Layer (`internal/handler/`)
- **gRPC handlers**: Handle gRPC requests and responses
- **HTTP handlers**: Auto-generated via gRPC-Gateway 
- **Consumer handlers**: Process events from message queue

### Repository Layer (`internal/repository/`)
- Data access using sqlc-generated type-safe SQL code
- PostgreSQL integration with connection pooling

## Services

The template includes two main services:

### Server (`cmd/server/`)
- **HTTP Server**: REST API on port 8080 (auto-generated from gRPC)
- **gRPC Server**: Native gRPC API on port 9090
- Serves both User and Product APIs with full CRUD operations

### Consumer (`cmd/consumer/`)
- Processes events from the message queue
- Handles `UserCreated`, `UserUpdated`, `ProductCreated`, etc.
- Demonstrates event-driven architecture patterns

## Event-Driven Architecture

- Uses **Watermill** with PostgreSQL as message broker
- Events are defined in `proto/event/v1/` using Protocol Buffers
- Automatic event generation using `voi-oss/protoc-gen-event`
- Events are published on entity creation/updates and consumed asynchronously

## Code Generation

The project uses several code generation tools:
- **buf**: Generates Go code from protobuf definitions
- **sqlc**: Generates type-safe Go code from SQL queries  
- **Atlas**: Manages database schema and migrations
- **protoc-gen-event**: Generates event handling code

## Database

- **Main DB**: PostgreSQL on port 5432 for application data
- **Message Queue DB**: Separate PostgreSQL on port 5433 for Watermill
- **Migrations**: Managed with Atlas in `db/migrations/`
- **Schema**: Current state in `db/schema.sql`
- **Queries**: SQL definitions in `db/queries/` with sqlc generation