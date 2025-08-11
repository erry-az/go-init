package watermill

import (
	"context"
	"database/sql"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watersql "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	publisher message.Publisher
	logger    watermill.LoggerAdapter
}

func NewPublisher(db *sql.DB, logger watermill.LoggerAdapter) (*Publisher, error) {
	publisher, err := watersql.NewPublisher(
		db,
		watersql.PublisherConfig{
			SchemaAdapter: watersql.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		publisher: publisher,
		logger:    logger,
	}, nil
}

func (p *Publisher) PublishProtoMessage(ctx context.Context, topic string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	watermillMsg := message.NewMessage(uuid.New().String(), data)
	watermillMsg.Metadata.Set("content-type", "application/x-protobuf")
	watermillMsg.Metadata.Set("timestamp", time.Now().Format(time.RFC3339))

	return p.publisher.Publish(topic, watermillMsg)
}

func (p *Publisher) Publish(ctx context.Context, topic string, data []byte, metadata map[string]string) error {
	watermillMsg := message.NewMessage(uuid.New().String(), data)
	
	for key, value := range metadata {
		watermillMsg.Metadata.Set(key, value)
	}
	
	watermillMsg.Metadata.Set("timestamp", time.Now().Format(time.RFC3339))

	return p.publisher.Publish(topic, watermillMsg)
}

func (p *Publisher) Close() error {
	return p.publisher.Close()
}