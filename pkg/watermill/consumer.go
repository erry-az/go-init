package watermill

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	subscriber *amqp.Subscriber
	router     *message.Router
	config     *Config
	logger     watermill.LoggerAdapter
}

type Handler func(ctx context.Context, msg proto.Message) error

func NewConsumer(config *Config, logger watermill.LoggerAdapter) (*Consumer, error) {
	amqpConfig := amqp.NewDurablePubSubConfig(config.AMQPURL, nil)
	amqpConfig.Marshaler = amqp.CorrelatingMarshaler{}

	if config.Exchange != "" {
		amqpConfig.Exchange = amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return config.Exchange
			},
			Type:    config.ExchangeType,
			Durable: config.Durable,
		}
	}

	if config.QueueConfig.GenerateName != nil {
		amqpConfig.Queue = amqp.QueueConfig{
			GenerateName: config.QueueConfig.GenerateName,
			Durable:      config.QueueConfig.Durable,
			AutoDelete:   config.QueueConfig.AutoDelete,
			Exclusive:    config.QueueConfig.Exclusive,
			NoWait:       config.QueueConfig.NoWait,
			Arguments:    config.QueueConfig.Arguments,
		}
	}

	subscriber, err := amqp.NewSubscriber(amqpConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	return &Consumer{
		subscriber: subscriber,
		router:     router,
		config:     config,
		logger:     logger,
	}, nil
}

func (c *Consumer) Subscribe(topic string, handler Handler, msgType proto.Message) error {
	c.router.AddNoPublisherHandler(
		topic+"_handler",
		topic,
		c.subscriber,
		func(msg *message.Message) error {
			event := proto.Clone(msgType)
			if err := proto.Unmarshal(msg.Payload, event); err != nil {
				c.logger.Error("Failed to unmarshal message", err, watermill.LogFields{
					"topic": topic,
					"uuid":  msg.UUID,
				})
				return err
			}

			ctx := context.Background()
			return handler(ctx, event)
		},
	)

	return nil
}

func (c *Consumer) Run(ctx context.Context) error {
	return c.router.Run(ctx)
}

func (c *Consumer) Close() error {
	if err := c.router.Close(); err != nil {
		return fmt.Errorf("failed to close router: %w", err)
	}

	if err := c.subscriber.Close(); err != nil {
		return fmt.Errorf("failed to close subscriber: %w", err)
	}

	return nil
}
