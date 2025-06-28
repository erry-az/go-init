package rabbitmq

import (
	"context"
	"log"

	apiv1 "github.com/erry-az/go-init/api/v1"
)

// ExampleUsage demonstrates how to use the new RabbitMQ client
func ExampleUsage() error {
	ctx := context.Background()

	// Create a new client
	client, err := NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer client.Close()

	// Setup default mappings (this handles exchange, queue, and binding setup)
	if err := SetupDefaultMappings(client); err != nil {
		return err
	}

	// Example 1: Publishing events - now using standalone generic function
	userCreatedEvent := &apiv1.UserCreatedEvent{
		Id:        "123e4567-e89b-12d3-a456-426614174000",
		Name:      "John Doe",
		Email:     "john.doe@example.com",
		CreatedAt: "2024-05-21T10:00:00Z",
	}

	if err := Publish(ctx, client, userCreatedEvent); err != nil {
		log.Printf("Failed to publish UserCreatedEvent: %v", err)
		return err
	}

	// Example 2: Subscribing to events with type-safe handlers
	err = Subscribe(ctx, client, &apiv1.UserCreatedEvent{}, func(ctx context.Context, event *apiv1.UserCreatedEvent) error {
		log.Printf("Received UserCreatedEvent: ID=%s, Name=%s, Email=%s", 
			event.Id, event.Name, event.Email)
		
		// Your business logic here
		// For example, send welcome email, update analytics, etc.
		
		return nil
	})
	if err != nil {
		return err
	}

	err = Subscribe(ctx, client, &apiv1.UserUpdatedEvent{}, func(ctx context.Context, event *apiv1.UserUpdatedEvent) error {
		log.Printf("Received UserUpdatedEvent: ID=%s, Name=%s, Email=%s", 
			event.Id, event.Name, event.Email)
		
		// Your business logic here
		// For example, update search index, invalidate cache, etc.
		
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ExampleCustomMapping demonstrates how to create custom mappings
func ExampleCustomMapping() error {
	client, err := NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer client.Close()

	// Create a custom mapping using the builder pattern
	customMapping := NewMappingBuilder("UserCreatedEvent").
		WithExchange("custom_events", "topic").
		WithQueue("custom_user_created_queue").
		WithBinding("custom_events", "custom_user_created_queue", "custom.user.created").
		WithConsumer("custom_user_created_queue").
		WithPublishingKey("custom.user.created").
		Build()

	// Register the custom mapping
	if err := client.RegisterEventMapping(customMapping); err != nil {
		return err
	}

	// Now you can publish and subscribe using the custom mapping
	// The client will automatically use the correct exchange and routing key
	ctx := context.Background()
	
	event := &apiv1.UserCreatedEvent{
		Id:        "custom-123",
		Name:      "Jane Doe",
		Email:     "jane.doe@example.com",
		CreatedAt: "2024-05-21T11:00:00Z",
	}

	return Publish(ctx, client, event)
}

// ExampleAdvancedConfiguration demonstrates advanced configuration options
func ExampleAdvancedConfiguration() error {
	client, err := NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer client.Close()

	// Advanced mapping with custom configurations
	advancedMapping := &EventMapping{
		EventTypeName: "UserCreatedEvent",
		Exchange: ExchangeConfig{
			Name:       "advanced_events",
			Type:       "topic",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Args: map[string]interface{}{
				"x-message-ttl": 3600000, // 1 hour TTL
			},
		},
		Queue: QueueConfig{
			Name:       "advanced_user_queue",
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args: map[string]interface{}{
				"x-max-length":        1000,   // Max 1000 messages
				"x-message-ttl":       3600000, // 1 hour TTL
				"x-dead-letter-exchange": "dlx", // Dead letter exchange
			},
		},
		Binding: BindingConfig{
			Exchange:   "advanced_events",
			Queue:      "advanced_user_queue",
			RoutingKey: "user.*",
			NoWait:     false,
			Args:       nil,
		},
		Consumer: ConsumerConfig{
			Queue:     "advanced_user_queue",
			Consumer:  "advanced-consumer",
			AutoAck:   false, // Manual acknowledgment for reliability
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
			Args:      nil,
		},
		PublishingKey: "user.created",
	}

	return client.RegisterEventMapping(advancedMapping)
}

// ExampleWithHelperFunctions demonstrates using helper functions for common patterns
func ExampleWithHelperFunctions() error {
	client, err := NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}
	defer client.Close()

	// Using helper functions to create mappings
	userCreatedMapping := CreateUserCreatedEventMapping("my_events", "my_user_created_queue", "user.created")
	userUpdatedMapping := CreateUserUpdatedEventMapping("my_events", "my_user_updated_queue", "user.updated")

	// Register the mappings
	if err := client.RegisterEventMapping(userCreatedMapping); err != nil {
		return err
	}
	if err := client.RegisterEventMapping(userUpdatedMapping); err != nil {
		return err
	}

	ctx := context.Background()

	// Publish events
	userCreatedEvent := &apiv1.UserCreatedEvent{
		Id:        "helper-123",
		Name:      "Helper User",
		Email:     "helper@example.com",
		CreatedAt: "2024-05-21T12:00:00Z",
	}

	if err := Publish(ctx, client, userCreatedEvent); err != nil {
		return err
	}

	// Subscribe to events
	return Subscribe(ctx, client, &apiv1.UserCreatedEvent{}, func(ctx context.Context, event *apiv1.UserCreatedEvent) error {
		log.Printf("Helper function example - User created: %s", event.Name)
		return nil
	})
}