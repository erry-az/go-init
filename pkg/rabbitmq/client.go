package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"reflect"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

// ProtoMessage is an interface that all protobuf messages implement
type ProtoMessage interface {
	proto.Message
}

// EventHandler represents a handler function for a specific event type
type EventHandler[T ProtoMessage] func(ctx context.Context, event T) error

// ExchangeConfig defines the configuration for an exchange
type ExchangeConfig struct {
	Name       string
	Type       string // "topic", "direct", "fanout", "headers"
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

// QueueConfig defines the configuration for a queue
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

// BindingConfig defines the configuration for binding a queue to an exchange
type BindingConfig struct {
	Exchange   string
	Queue      string
	RoutingKey string
	NoWait     bool
	Args       amqp.Table
}

// ConsumerConfig defines the configuration for a consumer
type ConsumerConfig struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

// EventMapping defines the mapping between an event type and its RabbitMQ configuration
type EventMapping struct {
	EventTypeName string
	Exchange      ExchangeConfig
	Queue         QueueConfig
	Binding       BindingConfig
	Consumer      ConsumerConfig
	PublishingKey string
}

// Client represents a generic RabbitMQ client
type Client struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	mappings map[string]*EventMapping
}

// NewClient creates a new RabbitMQ client
func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &Client{
		conn:     conn,
		channel:  ch,
		mappings: make(map[string]*EventMapping),
	}, nil
}

// Close closes the channel and connection
func (c *Client) Close() error {
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

// RegisterEventMapping registers an event mapping and sets up the necessary infrastructure
func (c *Client) RegisterEventMapping(mapping *EventMapping) error {
	// Declare exchange
	err := c.channel.ExchangeDeclare(
		mapping.Exchange.Name,
		mapping.Exchange.Type,
		mapping.Exchange.Durable,
		mapping.Exchange.AutoDelete,
		mapping.Exchange.Internal,
		mapping.Exchange.NoWait,
		mapping.Exchange.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", mapping.Exchange.Name, err)
	}

	// Declare queue
	_, err = c.channel.QueueDeclare(
		mapping.Queue.Name,
		mapping.Queue.Durable,
		mapping.Queue.AutoDelete,
		mapping.Queue.Exclusive,
		mapping.Queue.NoWait,
		mapping.Queue.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", mapping.Queue.Name, err)
	}

	// Bind queue to exchange
	err = c.channel.QueueBind(
		mapping.Binding.Queue,
		mapping.Binding.RoutingKey,
		mapping.Binding.Exchange,
		mapping.Binding.NoWait,
		mapping.Binding.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w", 
			mapping.Binding.Queue, mapping.Binding.Exchange, err)
	}

	// Store mapping for later use
	c.mappings[mapping.EventTypeName] = mapping

	return nil
}

// getEventTypeName returns the type name of the event for mapping purposes
func getEventTypeName(event ProtoMessage) string {
	return reflect.TypeOf(event).Elem().Name()
}

// Publish publishes an event using the registered mapping
func Publish[T ProtoMessage](ctx context.Context, client *Client, event T) error {
	eventType := getEventTypeName(event)
	
	mapping, exists := client.mappings[eventType]
	if !exists {
		return fmt.Errorf("no mapping registered for event type %s", eventType)
	}

	// Marshal the event
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event %s: %w", eventType, err)
	}

	// Publish the message
	return client.channel.PublishWithContext(
		ctx,
		mapping.Exchange.Name,
		mapping.PublishingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/protobuf",
			Body:        data,
		},
	)
}

// Subscribe subscribes to events using the registered mapping and handler
func Subscribe[T ProtoMessage](ctx context.Context, client *Client, eventType T, handler EventHandler[T]) error {
	typeName := getEventTypeName(eventType)
	
	mapping, exists := client.mappings[typeName]
	if !exists {
		return fmt.Errorf("no mapping registered for event type %s", typeName)
	}

	// Start consuming
	msgs, err := client.channel.Consume(
		mapping.Consumer.Queue,
		mapping.Consumer.Consumer,
		mapping.Consumer.AutoAck,
		mapping.Consumer.Exclusive,
		mapping.Consumer.NoLocal,
		mapping.Consumer.NoWait,
		mapping.Consumer.Args,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer for %s: %w", typeName, err)
	}

	// Process messages in a goroutine
	go func() {
		for msg := range msgs {
			// Create a new instance of the event type
			event := reflect.New(reflect.TypeOf(eventType).Elem()).Interface().(T)
			
			// Unmarshal the message
			err := proto.Unmarshal(msg.Body, event)
			if err != nil {
				log.Printf("Error unmarshaling message for %s: %v", typeName, err)
				if !mapping.Consumer.AutoAck {
					msg.Nack(false, true) // Requeue on unmarshal error
				}
				continue
			}

			// Handle the event
			err = handler(ctx, event)
			if err != nil {
				log.Printf("Error handling event %s: %v", typeName, err)
				if !mapping.Consumer.AutoAck {
					msg.Nack(false, true) // Requeue on handler error
				}
			} else {
				if !mapping.Consumer.AutoAck {
					msg.Ack(false) // Acknowledge successful processing
				}
			}
		}
	}()

	return nil
}