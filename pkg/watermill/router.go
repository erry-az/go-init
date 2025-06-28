package watermill

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"google.golang.org/protobuf/proto"
)

type EventRouter struct {
	router     *message.Router
	publisher  *amqp.Publisher
	subscriber *amqp.Subscriber
	config     *Config
	logger     watermill.LoggerAdapter
}

func NewEventRouter(config *Config, logger watermill.LoggerAdapter) (*EventRouter, error) {
	amqpConfig := amqp.NewDurablePubSubConfig(config.AMQPURL, nil)
	
	if config.Exchange != "" {
		amqpConfig.Exchange = amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return config.Exchange
			},
			Type:    config.ExchangeType,
			Durable: config.Durable,
		}
	}
	
	publisher, err := amqp.NewPublisher(amqpConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}
	
	subscriber, err := amqp.NewSubscriber(amqpConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}
	
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}
	
	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: middleware.DefaultInitialInterval,
			Logger:          logger,
		}.Middleware,
		middleware.Recoverer,
	)
	
	return &EventRouter{
		router:     router,
		publisher:  publisher,
		subscriber: subscriber,
		config:     config,
		logger:     logger,
	}, nil
}

func (r *EventRouter) AddHandler(handlerName, topic string, handler Handler, msgType proto.Message) {
	r.router.AddNoPublisherHandler(
		handlerName,
		topic,
		r.subscriber,
		r.wrapHandler(handler, msgType),
	)
}

func (r *EventRouter) wrapHandler(handler Handler, msgType proto.Message) message.HandlerFunc {
	return func(msg *message.Message) error {
		event := proto.Clone(msgType)
		if err := proto.Unmarshal(msg.Payload, event); err != nil {
			r.logger.Error("Failed to unmarshal message", err, watermill.LogFields{
				"uuid":         msg.UUID,
				"content-type": msg.Metadata.Get("content-type"),
			})
			return err
		}
		
		ctx := msg.Context()
		return handler(ctx, event)
	}
}

func (r *EventRouter) Publish(topic string, event proto.Message) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	msg := message.NewMessage(watermill.NewUUID(), data)
	msg.Metadata.Set("content-type", "application/protobuf")
	msg.Metadata.Set("type", string(proto.MessageName(event)))
	
	return r.publisher.Publish(topic, msg)
}

func (r *EventRouter) PublishWithContext(ctx context.Context, topic string, event proto.Message) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	msg := message.NewMessage(watermill.NewUUID(), data)
	msg.Metadata.Set("content-type", "application/protobuf")
	msg.Metadata.Set("type", string(proto.MessageName(event)))
	msg.SetContext(ctx)
	
	return r.publisher.Publish(topic, msg)
}

func (r *EventRouter) Run(ctx context.Context) error {
	return r.router.Run(ctx)
}

func (r *EventRouter) Close() error {
	if err := r.router.Close(); err != nil {
		return fmt.Errorf("failed to close router: %w", err)
	}
	
	if err := r.publisher.Close(); err != nil {
		return fmt.Errorf("failed to close publisher: %w", err)
	}
	
	if err := r.subscriber.Close(); err != nil {
		return fmt.Errorf("failed to close subscriber: %w", err)
	}
	
	return nil
}