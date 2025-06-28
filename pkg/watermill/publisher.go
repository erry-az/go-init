package watermill

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	publisher *amqp.Publisher
	config    *Config
	logger    watermill.LoggerAdapter
}

func NewPublisher(config *Config, logger watermill.LoggerAdapter) (*Publisher, error) {
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

	return &Publisher{
		publisher: publisher,
		config:    config,
		logger:    logger,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, topic string, event proto.Message) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := message.NewMessage(watermill.NewUUID(), data)
	msg.Metadata.Set("content-type", "application/protobuf")
	msg.Metadata.Set("type", string(proto.MessageName(event)))

	if err := p.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (p *Publisher) Close() error {
	return p.publisher.Close()
}
