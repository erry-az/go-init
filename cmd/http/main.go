package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	apiv1 "github.com/erry-az/go-init/api/v1"
	"github.com/erry-az/go-init/db/sqlc"
	"github.com/erry-az/go-init/internal/repository"
	"github.com/erry-az/go-init/internal/server"
	"github.com/erry-az/go-init/internal/service"
	"github.com/erry-az/go-init/pkg/rabbitmq"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	// Connect to PostgreSQL
	connString := "postgres://postgres:postgres@localhost:5432/go_init_db?sslmode=disable"
	pgxPool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgxPool.Close()

	// Create repositories
	queries := sqlc.New(pgxPool)
	repo := repository.NewPostgresRepository(queries)

	// Connect to RabbitMQ
	publisher, err := rabbitmq.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer publisher.Close()

	// Create services
	userService := service.NewUserService(repo, publisher)

	// Create and start gRPC server
	grpcServer := server.NewGRPCServer(userService, 9000)
	if err := grpcServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Create and start HTTP server
	httpServer := server.NewHttpServer(userService, 8080)
	if err := httpServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Setup RabbitMQ consumer (as an example)
	consumer, err := rabbitmq.NewConsumer("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()

	// Register example handler
	err = consumer.RegisterUserCreatedHandler(ctx, func(ctx context.Context, event *apiv1.UserCreatedEvent) error {
		log.Printf("Received UserCreatedEvent: %v", event)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to register user created handler: %v", err)
	}

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := grpcServer.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}

	log.Println("Stopping HTTP server...")
	if err := httpServer.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop HTTP server: %v", err)
	}

	log.Println("Server stopped")
}
