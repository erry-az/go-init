package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	apiv1 "github.com/erry-az/go-init/api/v1"
	"github.com/erry-az/go-init/pkg/watermill"
	"google.golang.org/protobuf/proto"
)

func main() {
	logger := watermill.NewStdLogger(false, false)

	// Create Watermill configuration
	config := watermill.DefaultConfig("amqp://guest:guest@localhost:5672/")
	config.Exchange = "user_events"
	config.ExchangeType = "topic"

	// Create event router
	router, err := watermill.NewEventRouter(config, logger)
	if err != nil {
		log.Fatalf("Failed to create event router: %v", err)
	}
	defer router.Close()

	// Register handlers
	router.AddHandler(
		"user_created_handler",
		"user.created",
		handleUserCreated,
		&apiv1.UserCreatedEvent{},
	)

	router.AddHandler(
		"user_updated_handler",
		"user.updated",
		handleUserUpdated,
		&apiv1.UserUpdatedEvent{},
	)

	// Run router
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	log.Println("Consumer started, waiting for messages...")
	if err := router.Run(ctx); err != nil {
		log.Fatalf("Failed to run router: %v", err)
	}

	log.Println("Consumer stopped")
}

func handleUserCreated(ctx context.Context, event proto.Message) error {
	userCreated := event.(*apiv1.UserCreatedEvent)
	log.Printf("Received UserCreatedEvent: ID=%s, Name=%s, Email=%s",
		userCreated.Id, userCreated.Name, userCreated.Email)

	// Add your business logic here
	// For example: send welcome email, update analytics, etc.

	return nil
}

func handleUserUpdated(ctx context.Context, event proto.Message) error {
	userUpdated := event.(*apiv1.UserUpdatedEvent)
	log.Printf("Received UserUpdatedEvent: ID=%s, Name=%s, Email=%s",
		userUpdated.Id, userUpdated.Name, userUpdated.Email)

	// Add your business logic here
	// For example: sync with external systems, update cache, etc.

	return nil
}
