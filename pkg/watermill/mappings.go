package watermill

import (
	"reflect"
	
	"google.golang.org/protobuf/proto"
)

type EventMapping struct {
	Topic        string
	RoutingKey   string
	QueueName    string
	EventType    reflect.Type
	ExchangeName string
}

var DefaultMappings = map[reflect.Type]EventMapping{}

func RegisterMapping(msgType proto.Message, mapping EventMapping) {
	eventType := reflect.TypeOf(msgType)
	mapping.EventType = eventType
	DefaultMappings[eventType] = mapping
}

func GetMapping(msgType proto.Message) (EventMapping, bool) {
	mapping, ok := DefaultMappings[reflect.TypeOf(msgType)]
	return mapping, ok
}

func GetTopicForEvent(event proto.Message) string {
	if mapping, ok := GetMapping(event); ok {
		return mapping.Topic
	}
	return ""
}

type MappingBuilder struct {
	mapping EventMapping
}

func NewMappingBuilder() *MappingBuilder {
	return &MappingBuilder{
		mapping: EventMapping{
			ExchangeName: "events",
		},
	}
}

func (b *MappingBuilder) WithTopic(topic string) *MappingBuilder {
	b.mapping.Topic = topic
	return b
}

func (b *MappingBuilder) WithRoutingKey(key string) *MappingBuilder {
	b.mapping.RoutingKey = key
	return b
}

func (b *MappingBuilder) WithQueueName(name string) *MappingBuilder {
	b.mapping.QueueName = name
	return b
}

func (b *MappingBuilder) WithExchange(name string) *MappingBuilder {
	b.mapping.ExchangeName = name
	return b
}

func (b *MappingBuilder) Build() EventMapping {
	return b.mapping
}

func init() {
	// Register default mappings for known events
	// These will be populated when the events package is imported
}