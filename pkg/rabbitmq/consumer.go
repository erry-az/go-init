package rabbitmq

import (
	"context"
	"fmt"
	apiv1 "github.com/erry-az/go-init/api/v1"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

// MessageHandler is a function that handles a RabbitMQ message
type MessageHandler func(context.Context, []byte) error

// Consumer handles consuming messages from RabbitMQ
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queues  map[string]*amqp.Queue
}

// NewConsumer creates a new RabbitMQ consumer
func NewConsumer(url string) (*Consumer, error) {
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

	return &Consumer{
		conn:    conn,
		channel: ch,
		queues:  make(map[string]*amqp.Queue),
	}, nil
}

// Close closes the channel and connection
func (c *Consumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

// RegisterUserCreatedHandler registers a handler for UserCreatedEvent
func (c *Consumer) RegisterUserCreatedHandler(ctx context.Context, handler func(context.Context, *apiv1.UserCreatedEvent) error) error {
	return c.registerHandler(ctx, "user_events", "user.created", "user_created_queue", func(ctx context.Context, data []byte) error {
		var event apiv1.UserCreatedEvent
		if err := proto.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal UserCreatedEvent: %w", err)
		}
		return handler(ctx, &event)
	})
}

// RegisterUserUpdatedHandler registers a handler for UserUpdatedEvent
func (c *Consumer) RegisterUserUpdatedHandler(ctx context.Context, handler func(context.Context, *apiv1.UserUpdatedEvent) error) error {
	return c.registerHandler(ctx, "user_events", "user.updated", "user_updated_queue", func(ctx context.Context, data []byte) error {
		var event apiv1.UserUpdatedEvent
		if err := proto.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("failed to unmarshal UserUpdatedEvent: %w", err)
		}
		return handler(ctx, &event)
	})
}

// registerHandler registers a handler for a specific exchange, routing key, and queue
func (c *Consumer) registerHandler(ctx context.Context, exchange, routingKey, queueName string, handler MessageHandler) error {
	// Declare a queue
	q, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	c.queues[queueName] = &q

	// Bind the queue to the exchange
	err = c.channel.QueueBind(
		q.Name,     // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %w", err)
	}

	// Start consuming
	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Process messages
	go func() {
		for msg := range msgs {
			err := handler(ctx, msg.Body)
			if err != nil {
				log.Printf("Error handling message: %v", err)
				msg.Nack(false, true) // Negative acknowledgement, requeue
			} else {
				msg.Ack(false) // Acknowledgement
			}
		}
	}()

	return nil
}
