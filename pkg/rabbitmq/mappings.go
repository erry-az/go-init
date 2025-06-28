package rabbitmq

// SetupDefaultMappings sets up all default mappings for a client
func SetupDefaultMappings(client *Client) error {
	// Register UserCreatedEvent mapping
	if err := client.RegisterEventMapping(UserCreatedEventMapping()); err != nil {
		return err
	}

	// Register UserUpdatedEvent mapping
	if err := client.RegisterEventMapping(UserUpdatedEventMapping()); err != nil {
		return err
	}

	return nil
}

// UserCreatedEventMapping returns the mapping configuration for UserCreatedEvent
func UserCreatedEventMapping() *EventMapping {
	return &EventMapping{
		EventTypeName: "UserCreatedEvent",
		Exchange: ExchangeConfig{
			Name:       "user_events",
			Type:       "topic",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Args:       nil,
		},
		Queue: QueueConfig{
			Name:       "user_created_queue",
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		Binding: BindingConfig{
			Exchange:   "user_events",
			Queue:      "user_created_queue",
			RoutingKey: "user.created",
			NoWait:     false,
			Args:       nil,
		},
		Consumer: ConsumerConfig{
			Queue:     "user_created_queue",
			Consumer:  "",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
			Args:      nil,
		},
		PublishingKey: "user.created",
	}
}

// UserUpdatedEventMapping returns the mapping configuration for UserUpdatedEvent
func UserUpdatedEventMapping() *EventMapping {
	return &EventMapping{
		EventTypeName: "UserUpdatedEvent",
		Exchange: ExchangeConfig{
			Name:       "user_events",
			Type:       "topic",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Args:       nil,
		},
		Queue: QueueConfig{
			Name:       "user_updated_queue",
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		Binding: BindingConfig{
			Exchange:   "user_events",
			Queue:      "user_updated_queue",
			RoutingKey: "user.updated",
			NoWait:     false,
			Args:       nil,
		},
		Consumer: ConsumerConfig{
			Queue:     "user_updated_queue",
			Consumer:  "",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
			Args:      nil,
		},
		PublishingKey: "user.updated",
	}
}

// Builder pattern for creating custom mappings
type MappingBuilder struct {
	mapping *EventMapping
}

// NewMappingBuilder creates a new mapping builder for an event type
func NewMappingBuilder(eventTypeName string) *MappingBuilder {
	return &MappingBuilder{
		mapping: &EventMapping{
			EventTypeName: eventTypeName,
		},
	}
}

// WithExchange sets the exchange configuration
func (b *MappingBuilder) WithExchange(name, exchangeType string) *MappingBuilder {
	b.mapping.Exchange = ExchangeConfig{
		Name:       name,
		Type:       exchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}
	return b
}

// WithQueue sets the queue configuration
func (b *MappingBuilder) WithQueue(name string) *MappingBuilder {
	b.mapping.Queue = QueueConfig{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
	return b
}

// WithBinding sets the binding configuration
func (b *MappingBuilder) WithBinding(exchange, queue, routingKey string) *MappingBuilder {
	b.mapping.Binding = BindingConfig{
		Exchange:   exchange,
		Queue:      queue,
		RoutingKey: routingKey,
		NoWait:     false,
		Args:       nil,
	}
	return b
}

// WithConsumer sets the consumer configuration
func (b *MappingBuilder) WithConsumer(queue string) *MappingBuilder {
	b.mapping.Consumer = ConsumerConfig{
		Queue:     queue,
		Consumer:  "",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	}
	return b
}

// WithPublishingKey sets the publishing routing key
func (b *MappingBuilder) WithPublishingKey(key string) *MappingBuilder {
	b.mapping.PublishingKey = key
	return b
}

// WithCustomExchange allows setting custom exchange configuration
func (b *MappingBuilder) WithCustomExchange(config ExchangeConfig) *MappingBuilder {
	b.mapping.Exchange = config
	return b
}

// WithCustomQueue allows setting custom queue configuration
func (b *MappingBuilder) WithCustomQueue(config QueueConfig) *MappingBuilder {
	b.mapping.Queue = config
	return b
}

// WithCustomConsumer allows setting custom consumer configuration
func (b *MappingBuilder) WithCustomConsumer(config ConsumerConfig) *MappingBuilder {
	b.mapping.Consumer = config
	return b
}

// Build returns the configured mapping
func (b *MappingBuilder) Build() *EventMapping {
	return b.mapping
}

// Helper functions to create mappings for specific protobuf types

// CreateUserCreatedEventMapping creates a mapping for UserCreatedEvent with custom configuration
func CreateUserCreatedEventMapping(exchangeName, queueName, routingKey string) *EventMapping {
	return NewMappingBuilder("UserCreatedEvent").
		WithExchange(exchangeName, "topic").
		WithQueue(queueName).
		WithBinding(exchangeName, queueName, routingKey).
		WithConsumer(queueName).
		WithPublishingKey(routingKey).
		Build()
}

// CreateUserUpdatedEventMapping creates a mapping for UserUpdatedEvent with custom configuration
func CreateUserUpdatedEventMapping(exchangeName, queueName, routingKey string) *EventMapping {
	return NewMappingBuilder("UserUpdatedEvent").
		WithExchange(exchangeName, "topic").
		WithQueue(queueName).
		WithBinding(exchangeName, queueName, routingKey).
		WithConsumer(queueName).
		WithPublishingKey(routingKey).
		Build()
}
