package watmil

import (
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watersql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
)

// NewPublisher creates a new event bus using pgxpool.Pool for database operations.
// The pool is converted to *sql.DB using stdlib connector for watermill-sql compatibility.
func NewPublisher(pool *pgxpool.Pool, logger watermill.LoggerAdapter) (*cqrs.EventBus, error) {
	publisher, err := watersql.NewPublisher(
		stdlib.OpenDBFromPool(pool),
		watersql.PublisherConfig{
			SchemaAdapter:        watersql.DefaultPostgreSQLSchema{},
			AutoInitializeSchema: true,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	tracePropagation := wotelfloss.NewTracePropagatingPublisherDecorator(publisher)

	eventBus, err := cqrs.NewEventBusWithConfig(wotel.NewPublisherDecorator(tracePropagation), cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return generateEventTopic(params.EventName), nil
		},
		OnPublish: func(params cqrs.OnEventSendParams) error {
			logger.Info("Publishing event", watermill.LogFields{
				"event_name": params.EventName,
			})

			params.Message.Metadata.Set("published_at", time.Now().Format(time.RFC3339))

			return nil
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		Logger: logger,
	})

	return eventBus, nil
}

func generateEventTopic(eventName string) string {
	return "events." + eventName
}
