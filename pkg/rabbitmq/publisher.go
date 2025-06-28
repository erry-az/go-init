package rabbitmq

import (
	"context"
	"fmt"
	apiv1 "github.com/erry-az/go-init/api/v1"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

// Publisher handles publishing messages to RabbitMQ
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare exchanges
	err = ch.ExchangeDeclare(
		"user_events", // exchange name
		"topic",       // exchange type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

// Close closes the channel and connection
func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

// PublishUserCreated publishes a UserCreatedEvent
func (p *Publisher) PublishUserCreated(ctx context.Context, event *apiv1.UserCreatedEvent) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal UserCreatedEvent: %w", err)
	}

	return p.publishMessage(ctx, "user_events", "user.created", data)
}

// PublishUserUpdated publishes a UserUpdatedEvent
func (p *Publisher) PublishUserUpdated(ctx context.Context, event *apiv1.UserUpdatedEvent) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal UserUpdatedEvent: %w", err)
	}

	return p.publishMessage(ctx, "user_events", "user.updated", data)
}

// publishMessage publishes a message to an exchange with a routing key
func (p *Publisher) publishMessage(ctx context.Context, exchange, routingKey string, body []byte) error {
	return p.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/protobuf",
			Body:        body,
		},
	)
}
