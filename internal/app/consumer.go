package app

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/erry-az/go-sample/config"
	"github.com/erry-az/go-sample/internal/handler/consumer"
	"github.com/erry-az/go-sample/pkg/watmil"
)

// ConsumerApp represents the consumer application
type ConsumerApp struct {
	ProductConsumer *consumer.ProductConsumer
	UserConsumer    *consumer.UserConsumer
	Subscriber      *watmil.Subscriber
}

// NewConsumerApp creates a new consumer application with all dependencies
func NewConsumerApp(cfg *config.Config) (*ConsumerApp, error) {
	// Create consumers
	productConsumer := consumer.NewProductConsumer()
	userConsumer := consumer.NewUserConsumer()

	// Create standard SQL connection for Watermill
	sqlDB, err := sql.Open("pgx", cfg.Databases.PgMqUrl)
	if err != nil {
		slog.Error("Failed to connect to SQL database", slog.Any("error", err))
		return nil, err
	}

	// Test SQL database connection
	if err := sqlDB.Ping(); err != nil {
		slog.Error("Failed to ping SQL database", slog.Any("error", err))
		sqlDB.Close()
		return nil, err
	}

	logger := watermill.NewSlogLogger(slog.Default())

	subscriber, err := watmil.NewSubscriber(sqlDB, logger,
		cfg.Consumers.Retry.MiddlewareRetry(logger).Middleware)
	if err != nil {
		slog.Error("Failed to subscribe to SQL database", slog.Any("error", err))
		sqlDB.Close()
		return nil, err
	}

	return &ConsumerApp{
		ProductConsumer: productConsumer,
		UserConsumer:    userConsumer,
		Subscriber:      subscriber,
	}, nil
}

// Run starts the consumer application
func (app *ConsumerApp) Run(ctx context.Context) error {
	err := app.Subscriber.RegisterHandlers(
		app.ProductConsumer.AddHandlers,
		app.UserConsumer.AddHandlers,
	)
	if err != nil {
		slog.Error("Failed to register handlers", slog.Any("error", err))
		return err
	}

	return app.Subscriber.Run(ctx)
}