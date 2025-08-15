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

type IntHandler interface {
	Handle(*message.Message) error
	Topic() string
}

type ProtoMessageHandler func(ctx context.Context, msg proto.Message) error

func NewSubscriber(db *sql.DB, logger watermill.LoggerAdapter) (*Subscriber, error) {
	subscriber, err := watersql.NewSubscriber(
		db,
		watersql.SubscriberConfig{
			SchemaAdapter:    watersql.DefaultPostgreSQLSchema{},
			OffsetsAdapter:   watersql.DefaultPostgreSQLOffsetsAdapter{},
			PollInterval:     time.Second,
			InitializeSchema: true,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	router.AddMiddleware()

	return &Subscriber{
		subscriber: subscriber,
		logger:     logger,
	}, nil
}

func (s *Subscriber) Subscribe(ctx context.Context, implHandler IntHandler) error {
	messages, err := s.subscriber.Subscribe(ctx, implHandler.Topic())
	if err != nil {
		return err
	}

	go func() {
		for msg := range messages {
			err := implHandler.Handle(msg)
			if err != nil {
				s.logger.Error("Failed to handle message", err, watermill.LogFields{
					"topic":      implHandler.Topic(),
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
