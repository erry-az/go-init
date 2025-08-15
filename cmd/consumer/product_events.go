package main

import (
	"context"
	"database/sql"
	"log"
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
	logger := watermillbase.NewStdLogger(false, false)

	// Create subscriber
	subscriber, err := watermill.NewSubscriber(db, logger)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to product events using generated handlers
	err = subscriber.Subscribe(
		ctx,
		eventv1.ProductCreatedEventHandler(consumer.HandleProductCreated),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to product created events: %v", err)
	}

	err = subscriber.Subscribe(
		ctx,
		eventv1.ProductUpdatedEventHandler(consumer.HandleProductUpdated),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to product updated events: %v", err)
	}

	err = subscriber.Subscribe(
		ctx,
		eventv1.ProductDeletedEventHandler(consumer.HandleProductDeleted),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to product deleted events: %v", err)
	}

	err = subscriber.Subscribe(
		ctx,
		eventv1.ProductPriceChangedEventHandler(consumer.HandleProductPriceChanged),
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to product price changed events: %v", err)
	}

	log.Println("Product events consumer started. Press Ctrl+C to exit...")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down product events consumer...")
	cancel()
	subscriber.Close()
}
