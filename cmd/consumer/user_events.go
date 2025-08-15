package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	watermillbase "github.com/ThreeDotsLabs/watermill"
	"github.com/erry-az/go-init/config"
	"github.com/erry-az/go-init/internal/handler/consumer"
	"github.com/erry-az/go-init/pkg/watermill"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create database connection
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create logger
	logger := watermillbase.NewSlogLogger(slog.Default())
	// Create subscriber
	subscriber, err := watermill.NewSubscriber(db, logger)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to user events using generated handlers
	err = subscriber.Subscribe(
		ctx,
		eventv1.UserCreatedEventHandler(consumer.HandleUserCreated),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to user created events: %v", err)
	}

	err = subscriber.Subscribe(
		ctx,
		eventv1.UserUpdatedEventHandler(consumer.HandleUserUpdated),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to user updated events: %v", err)
	}

	err = subscriber.Subscribe(
		ctx,
		eventv1.UserDeletedEventHandler(consumer.HandleUserDeleted),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to user deleted events: %v", err)
	}

	log.Println("User events consumer started. Press Ctrl+C to exit...")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down user events consumer...")
	cancel()
	subscriber.Close()
}
