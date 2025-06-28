# RabbitMQ Generic Client

A sophisticated RabbitMQ client that provides type-safe event publishing and consuming using Go generics. It maps exchanges, channels, queues, and consumers to gRPC-generated protobuf messages.

## Features

- **Type Safety**: Full type safety using Go generics with protobuf messages
- **Automatic Mapping**: Maps protobuf message types to RabbitMQ infrastructure
- **Configurable**: Flexible configuration for exchanges, queues, bindings, and consumers
- **Builder Pattern**: Easy-to-use builder pattern for custom mappings
- **Default Mappings**: Pre-configured mappings for common use cases
- **Error Handling**: Robust error handling with automatic retries and dead letter queues

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    apiv1 "github.com/erry-az/go-init/api/v1"
    "github.com/erry-az/go-init/pkg/rabbitmq"
)

func main() {
    ctx := context.Background()
    
    // Create client
    client, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Setup default mappings
    if err := rabbitmq.SetupDefaultMappings(client); err != nil {
        log.Fatal(err)
    }
    
    // Publish an event - using standalone generic function
    event := &apiv1.UserCreatedEvent{
        Id:        "123e4567-e89b-12d3-a456-426614174000",
        Name:      "John Doe",
        Email:     "john.doe@example.com",
        CreatedAt: "2024-05-21T10:00:00Z",
    }
    
    if err := rabbitmq.Publish(ctx, client, event); err != nil {
        log.Printf("Failed to publish: %v", err)
    }
    
    // Subscribe to events - using standalone generic function
    err = rabbitmq.Subscribe(ctx, client, &apiv1.UserCreatedEvent{}, func(ctx context.Context, event *apiv1.UserCreatedEvent) error {
        log.Printf("User created: %s (%s)", event.Name, event.Email)
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### Custom Mappings

```go
// Create custom mapping using builder pattern
customMapping := rabbitmq.NewMappingBuilder("UserCreatedEvent").
    WithExchange("my_events", "topic").
    WithQueue("my_user_queue").
    WithBinding("my_events", "my_user_queue", "user.created").
    WithConsumer("my_user_queue").
    WithPublishingKey("user.created").
    Build()

// Register the mapping
if err := client.RegisterEventMapping(customMapping); err != nil {
    log.Fatal(err)
}
```

### Advanced Configuration

```go
// Advanced mapping with custom configurations
advancedMapping := &rabbitmq.EventMapping{
    EventTypeName: "UserCreatedEvent",
    Exchange: rabbitmq.ExchangeConfig{
        Name:    "events",
        Type:    "topic",
        Durable: true,
        Args: map[string]interface{}{
            "x-message-ttl": 3600000, // 1 hour TTL
        },
    },
    Queue: rabbitmq.QueueConfig{
        Name:    "user_queue",
        Durable: true,
        Args: map[string]interface{}{
            "x-max-length": 1000,
            "x-dead-letter-exchange": "dlx",
        },
    },
    // ... other configurations
}
```

## Key Components

### EventMapping
Defines the complete mapping between a protobuf message type and RabbitMQ infrastructure:
- **EventTypeName**: String identifier for the protobuf message type
- **Exchange**: Exchange configuration
- **Queue**: Queue configuration  
- **Binding**: Queue-to-exchange binding configuration
- **Consumer**: Consumer configuration
- **PublishingKey**: Routing key for publishing

### Client
The main client that handles:
- Connection management
- Infrastructure setup (exchanges, queues, bindings)
- Event mapping registration

### Generic Functions
- **Publish[T]**: Type-safe publishing function
- **Subscribe[T]**: Type-safe subscription function with handlers

### Configuration Types
- **ExchangeConfig**: Exchange settings (name, type, durability, etc.)
- **QueueConfig**: Queue settings (name, durability, TTL, etc.)
- **BindingConfig**: Binding settings (routing keys, arguments)
- **ConsumerConfig**: Consumer settings (acknowledgments, exclusivity)

## Migration from Old Implementation

### Before (old implementation)
```go
// Publisher
publisher, err := rabbitmq.NewPublisher(url)
if err := publisher.PublishUserCreated(ctx, event); err != nil {
    // handle error
}

// Consumer  
consumer, err := rabbitmq.NewConsumer(url)
err = consumer.RegisterUserCreatedHandler(ctx, func(ctx context.Context, event *apiv1.UserCreatedEvent) error {
    // handle event
    return nil
})
```

### After (new implementation)
```go
// Single client for both publishing and consuming
client, err := rabbitmq.NewClient(url)
rabbitmq.SetupDefaultMappings(client)

// Publishing - type-safe, works with any protobuf message
if err := rabbitmq.Publish(ctx, client, event); err != nil {
    // handle error
}

// Consuming - type-safe handlers
err = rabbitmq.Subscribe(ctx, client, &apiv1.UserCreatedEvent{}, func(ctx context.Context, event *apiv1.UserCreatedEvent) error {
    // handle event
    return nil
})
```

### Legacy Compatibility

For existing code, legacy wrappers are provided:

```go
// Drop-in replacement for old Publisher
publisher, err := rabbitmq.NewLegacyPublisher(url)
publisher.PublishUserCreated(ctx, event)

// Drop-in replacement for old Consumer
consumer, err := rabbitmq.NewLegacyConsumer(url)
consumer.RegisterUserCreatedHandler(ctx, handler)
```

## Generic Functions vs Methods

Due to Go's limitation that struct methods cannot be generic, this implementation uses standalone generic functions:

- `Publish[T ProtoMessage](ctx, client, event)` instead of `client.Publish[T](ctx, event)`
- `Subscribe[T ProtoMessage](ctx, client, eventType, handler)` instead of `client.Subscribe[T](ctx, eventType, handler)`

This approach provides the same type safety while being compatible with Go's type system.

## Benefits

1. **Type Safety**: Compile-time type checking for event handlers
2. **Reduced Boilerplate**: No need to write separate publish/consume methods for each event type
3. **Flexibility**: Easy to configure exchanges, queues, and consumers
4. **Maintainability**: Single place to manage event mappings
5. **Extensibility**: Easy to add new event types by just defining mappings
6. **Error Handling**: Built-in error handling and retry mechanisms
7. **Backward Compatibility**: Legacy wrappers for existing code

## Default Event Mappings

The package provides default mappings for:
- **UserCreatedEvent**: `user_events` exchange, `user.created` routing key
- **UserUpdatedEvent**: `user_events` exchange, `user.updated` routing key

All default mappings use:
- Topic exchanges
- Durable queues and exchanges
- Manual acknowledgments for reliability
- Protobuf serialization

## Helper Functions

For common patterns, helper functions are provided:

```go
// Quick setup for standard events
userCreatedMapping := rabbitmq.CreateUserCreatedEventMapping("events", "user_queue", "user.created")
userUpdatedMapping := rabbitmq.CreateUserUpdatedEventMapping("events", "user_queue", "user.updated")
```