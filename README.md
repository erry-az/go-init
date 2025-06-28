# Go Microservice Template

A template for Go microservices using modern tools and best practices.

## Features

- Go modules with dependency management
- gRPC + gRPC-Gateway for HTTP/JSON and gRPC APIs
- OpenAPI/Swagger documentation generated from protobuf
- PostgreSQL database with sqlc for type-safe SQL
- RabbitMQ messaging with Protocol Buffers for event-driven architecture
- Protocol Buffer validation using buf.build's protovalidate
- Docker Compose for local development
- Nix shell for development environment
- Makefile for common tasks
- Mock generation using Uber's go-mock
- Database migrations with Goose
- To Be implemented: https://github.com/voi-oss/protoc-gen-event

## Requirements

- [Nix](https://nixos.org/download.html)
- [Rancher Desktop](https://rancherdesktop.io/) (for Docker management)

## Development

```shell
# Enter the development environment with Nix
nix-shell

# Start services (PostgreSQL, RabbitMQ)
make up

# Run database migrations
make db-migrate

# Generate code from proto files
make generate

# Run the service
make run
```

## Project Structure

```
.
├── api/                # API definitions and generated code
│   └── v1/             # API version 1 definitions
│       ├── events.proto    # Event definitions for message broker
│       └── service.proto   # Service definitions for gRPC
├── cmd/                # Application entry points
│   └── server/         # Main server application
├── db/                 # Database related code
│   ├── migrations/     # Goose database migrations
│   ├── queries/        # SQL queries for sqlc
│   └── sqlc/           # Generated sqlc code
├── docs/               # Generated documentation
│   └── v1/             # Generated Swagger/OpenAPI docs
├── internal/           # Private application code
│   ├── repository/     # Data access layer
│   ├── server/         # Server implementation
│   └── service/        # Business logic
├── mocks/              # Generated mocks for testing
├── pkg/                # Public libraries
│   └── rabbitmq/       # RabbitMQ integration
│       ├── consumer.go # Message consumer implementation
│       └── publisher.go# Message publisher implementation
└── scripts/            # Development and CI scripts
```

## Features in Detail

### API Definition

API definitions are written in Protocol Buffers (protobuf) and can be found in the `api/v1` directory. Service definitions include HTTP annotations for gRPC-Gateway, which generates REST endpoints from gRPC services.

### Events

The system uses RabbitMQ for event-driven architecture. Event definitions are in the `api/v1/events.proto` file and are serialized using Protocol Buffers.

### Code Generation

The project uses several code generation tools:
- `buf` generates Go code from protobuf definitions
- `sqlc` generates type-safe Go code from SQL queries
- `mockgen` generates mock implementations for testing

### Database

PostgreSQL is used as the database. Migrations are managed with Goose and can be found in the `db/migrations` directory. SQL queries are defined in the `db/queries` directory and type-safe Go code is generated using sqlc.

### Dependency Injection

The application uses a simple dependency injection pattern where services depend on repositories, which in turn depend on database connections.