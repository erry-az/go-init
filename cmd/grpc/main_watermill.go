package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	apiv1 "github.com/erry-az/go-init/api/v1"
	"github.com/erry-az/go-init/db/sqlc"
	"github.com/erry-az/go-init/internal/repository"
	"github.com/erry-az/go-init/internal/server"
	"github.com/erry-az/go-init/internal/service"
	"github.com/erry-az/go-init/pkg/watermill"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
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

	// Setup Watermill
	logger := watermill.NewStdLogger(false, false)
	config := watermill.DefaultConfig("amqp://guest:guest@localhost:5672/")
	config.Exchange = "user_events"
	config.ExchangeType = "topic"

	// Create event router
	router, err := watermill.NewEventRouter(config, logger)
	if err != nil {
		log.Fatalf("Failed to create event router: %v", err)
	}
	defer router.Close()

	// Register event handlers
	router.AddHandler(
		"user_created_handler",
		"user.created",
		func(ctx context.Context, event proto.Message) error {
			userCreated := event.(*apiv1.UserCreatedEvent)
			log.Printf("Received UserCreatedEvent: ID=%s, Name=%s, Email=%s", 
				userCreated.Id, userCreated.Name, userCreated.Email)
			return nil
		},
		&apiv1.UserCreatedEvent{},
	)

	// Start event router in background
	routerCtx, routerCancel := context.WithCancel(context.Background())
	defer routerCancel()
	
	go func() {
		if err := router.Run(routerCtx); err != nil {
			log.Printf("Event router error: %v", err)
		}
	}()

	// Create services with Watermill
	userService := service.NewUserServiceWithWatermill(repo, router)

	// Create and start gRPC server
	grpcServer := server.NewGRPCServerWithWatermill(userService, 9000)
	if err := grpcServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Create and start HTTP server
	httpServer := server.NewHttpServerWithWatermill(userService, 8080)
	if err := httpServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	
	// Stop event router
	routerCancel()
	
	// Stop gRPC server
	if err := grpcServer.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop server: %v", err)
	}

	// Stop HTTP server
	log.Println("Stopping HTTP server...")
	if err := httpServer.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop HTTP server: %v", err)
	}

	log.Println("Server stopped")
}