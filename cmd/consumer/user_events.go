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
	"github.com/erry-az/go-init/pkg/watermill"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/protobuf/proto"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
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

	// Subscribe to user events
	err = subscriber.SubscribeProto(
		ctx,
		watermill.TopicUserCreated,
		&eventv1.UserCreatedEvent{},
		handleUserCreated,
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to user created events: %v", err)
	}

	err = subscriber.SubscribeProto(
		ctx,
		watermill.TopicUserUpdated,
		&eventv1.UserUpdatedEvent{},
		handleUserUpdated,
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to user updated events: %v", err)
	}

	err = subscriber.SubscribeProto(
		ctx,
		watermill.TopicUserDeleted,
		&eventv1.UserDeletedEvent{},
		handleUserDeleted,
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

func handleUserCreated(ctx context.Context, msg proto.Message) error {
	userEvent := msg.(*eventv1.UserCreatedEvent)
	log.Printf("User created: ID=%s, Name=%s, Email=%s, EventID=%s, Source=%s",
		userEvent.User.Id,
		userEvent.User.Name,
		userEvent.User.Email,
		userEvent.EventId,
		userEvent.Data.Source,
	)

	// Here you could:
	// - Send welcome email
	// - Create user profile in another service
	// - Update analytics
	// - Log audit trail
	// - Access metadata: userEvent.Data.Metadata

	return nil
}

func handleUserUpdated(ctx context.Context, msg proto.Message) error {
	userEvent := msg.(*eventv1.UserUpdatedEvent)
	log.Printf("User updated: ID=%s, Name=%s, Email=%s, EventID=%s, Source=%s, ChangedFields=%v",
		userEvent.User.Id,
		userEvent.User.Name,
		userEvent.User.Email,
		userEvent.EventId,
		userEvent.Data.Source,
		userEvent.Data.ChangedFields,
	)

	// Here you could:
	// - Update cached user data
	// - Sync with external systems
	// - Update search indexes
	// - Access previous user: userEvent.Data.PreviousUser
	// - Access metadata: userEvent.Data.Metadata

	return nil
}

func handleUserDeleted(ctx context.Context, msg proto.Message) error {
	userEvent := msg.(*eventv1.UserDeletedEvent)
	log.Printf("User deleted: ID=%s, Name=%s, EventID=%s, Source=%s, Reason=%s",
		userEvent.User.Id,
		userEvent.User.Name,
		userEvent.EventId,
		userEvent.Data.Source,
		userEvent.Data.Reason,
	)

	// Here you could:
	// - Clean up user data
	// - Cancel subscriptions
	// - Archive user information
	// - Update analytics
	// - Access metadata: userEvent.Data.Metadata

	return nil
}
