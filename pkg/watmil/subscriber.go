package watmil

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watersql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	wotelfloss "github.com/dentech-floss/watermill-opentelemetry-go-extra/pkg/opentelemetry"
	wotel "github.com/voi-oss/watermill-opentelemetry/pkg/opentelemetry"
)

type Subscriber struct {
	router         *message.Router
	logger         watermill.LoggerAdapter
	eventProcessor *cqrs.EventProcessor
}

func NewSubscriber(db *sql.DB, logger watermill.LoggerAdapter, mid ...message.HandlerMiddleware) (*Subscriber, error) {
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	router.AddPlugin(plugin.SignalsHandler)
	router.AddMiddleware(middleware.Recoverer, wotelfloss.ExtractRemoteParentSpanContext(), wotel.Trace())
	router.AddMiddleware(mid...)

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return generateEventTopic(params.EventName), nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return watersql.NewSubscriber(
					db,
					watersql.SubscriberConfig{
						SchemaAdapter:    watersql.DefaultPostgreSQLSchema{},
						OffsetsAdapter:   watersql.DefaultPostgreSQLOffsetsAdapter{},
						InitializeSchema: true,
					},
					logger,
				)
			},
			OnHandle: func(params cqrs.EventProcessorOnHandleParams) error {
				start := time.Now()

				err := params.Handler.Handle(params.Message.Context(), params.Event)

				logger.Info("Event handled", watermill.LogFields{
					"event_name": params.EventName,
					"duration":   time.Since(start),
					"err":        err,
				})

				return err
			},
			Marshaler: cqrs.JSONMarshaler{
				GenerateName: cqrs.StructName,
			},
			Logger: logger,
		},
	)

	return &Subscriber{
		router:         router,
		logger:         logger,
		eventProcessor: eventProcessor,
	}, nil
}

func (s *Subscriber) RegisterHandlers(handlers ...func(eventProcessor *cqrs.EventProcessor) error) error {
	for _, handler := range handlers {
		if err := handler(s.eventProcessor); err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) Run(ctx context.Context) error {
	return s.router.Run(ctx)
}
