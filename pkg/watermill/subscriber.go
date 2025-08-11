package watermill

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watersql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"google.golang.org/protobuf/proto"
)

type Subscriber struct {
	subscriber message.Subscriber
	logger     watermill.LoggerAdapter
}

type ProtoMessageHandler func(ctx context.Context, msg proto.Message) error

func NewSubscriber(db *sql.DB, logger watermill.LoggerAdapter) (*Subscriber, error) {
	subscriber, err := watersql.NewSubscriber(
		db,
		watersql.SubscriberConfig{
			SchemaAdapter:  watersql.DefaultPostgreSQLSchema{},
			OffsetsAdapter: watersql.DefaultPostgreSQLOffsetsAdapter{},
			PollInterval:   time.Second,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		subscriber: subscriber,
		logger:     logger,
	}, nil
}

func (s *Subscriber) Subscribe(ctx context.Context, topic string, handler message.HandlerFunc) error {
	messages, err := s.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return err
	}

	go func() {
		for msg := range messages {
			_, err := handler(msg)
			if err != nil {
				s.logger.Error("Failed to handle message", err, watermill.LogFields{
					"topic":      topic,
					"message_id": msg.UUID,
				})
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()

	return nil
}

func (s *Subscriber) SubscribeProto(ctx context.Context, topic string, protoMsg proto.Message, handler ProtoMessageHandler) error {
	messages, err := s.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return err
	}

	go func() {
		for msg := range messages {
			// Clone the proto message to avoid concurrent access issues
			clonedMsg := proto.Clone(protoMsg)
			
			if err := proto.Unmarshal(msg.Payload, clonedMsg); err != nil {
				s.logger.Error("Failed to unmarshal proto message", err, watermill.LogFields{
					"topic":      topic,
					"message_id": msg.UUID,
				})
				msg.Nack()
				continue
			}

			if err := handler(ctx, clonedMsg); err != nil {
				s.logger.Error("Failed to handle proto message", err, watermill.LogFields{
					"topic":      topic,
					"message_id": msg.UUID,
				})
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()

	return nil
}

func (s *Subscriber) Close() error {
	return s.subscriber.Close()
}