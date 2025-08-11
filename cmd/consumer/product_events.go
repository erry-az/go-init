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
	logger := watermillbase.NewStdLogger(false, false)

	// Create subscriber
	subscriber, err := watermill.NewSubscriber(db, logger)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to product events
	err = subscriber.SubscribeProto(
		ctx,
		watermill.TopicProductCreated,
		&eventv1.ProductCreatedEvent{},
		handleProductCreated,
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to product created events: %v", err)
	}

	err = subscriber.SubscribeProto(
		ctx,
		watermill.TopicProductUpdated,
		&eventv1.ProductUpdatedEvent{},
		handleProductUpdated,
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to product updated events: %v", err)
	}

	err = subscriber.SubscribeProto(
		ctx,
		watermill.TopicProductDeleted,
		&eventv1.ProductDeletedEvent{},
		handleProductDeleted,
	)
	if err != nil {
		log.Fatalf("Failed to subscribe to product deleted events: %v", err)
	}

	err = subscriber.SubscribeProto(
		ctx,
		watermill.TopicProductPriceChanged,
		&eventv1.ProductPriceChangedEvent{},
		handleProductPriceChanged,
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

func handleProductCreated(ctx context.Context, msg proto.Message) error {
	productEvent := msg.(*eventv1.ProductCreatedEvent)
	log.Printf("Product created: ID=%s, Name=%s, Price=%s, EventID=%s, Source=%s",
		productEvent.Product.Id,
		productEvent.Product.Name,
		productEvent.Product.Price,
		productEvent.EventId,
		productEvent.Data.Source,
	)

	// Here you could:
	// - Update search index
	// - Sync with inventory system
	// - Update analytics
	// - Send notifications
	// - Access metadata: productEvent.Data.Metadata

	return nil
}

func handleProductUpdated(ctx context.Context, msg proto.Message) error {
	productEvent := msg.(*eventv1.ProductUpdatedEvent)
	log.Printf("Product updated: ID=%s, Name=%s, Price=%s, EventID=%s, Source=%s, ChangedFields=%v",
		productEvent.Product.Id,
		productEvent.Product.Name,
		productEvent.Product.Price,
		productEvent.EventId,
		productEvent.Data.Source,
		productEvent.Data.ChangedFields,
	)

	// Here you could:
	// - Update cached data
	// - Sync with external systems
	// - Update search indexes
	// - Access previous product: productEvent.Data.PreviousProduct
	// - Access metadata: productEvent.Data.Metadata

	return nil
}

func handleProductDeleted(ctx context.Context, msg proto.Message) error {
	productEvent := msg.(*eventv1.ProductDeletedEvent)
	log.Printf("Product deleted: ID=%s, Name=%s, EventID=%s, Source=%s, Reason=%s",
		productEvent.Product.Id,
		productEvent.Product.Name,
		productEvent.EventId,
		productEvent.Data.Source,
		productEvent.Data.Reason,
	)

	// Here you could:
	// - Remove from search index
	// - Clean up related data
	// - Update analytics
	// - Archive product information
	// - Access metadata: productEvent.Data.Metadata

	return nil
}

func handleProductPriceChanged(ctx context.Context, msg proto.Message) error {
	productEvent := msg.(*eventv1.ProductPriceChangedEvent)
	log.Printf("Product price changed: ID=%s, Name=%s, PreviousPrice=%s, NewPrice=%s, EventID=%s, Source=%s",
		productEvent.Product.Id,
		productEvent.Product.Name,
		productEvent.Data.PreviousPrice,
		productEvent.Data.NewPrice,
		productEvent.EventId,
		productEvent.Data.Source,
	)

	// Here you could:
	// - Update pricing alerts
	// - Recalculate recommendations
	// - Update analytics dashboards
	// - Send price change notifications
	// - Access metadata: productEvent.Data.Metadata

	return nil
}
