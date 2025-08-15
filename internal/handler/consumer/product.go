package consumer

import (
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
)

func HandleProductCreated(pe *eventv1.ProductCreatedEvent, m *message.Message) error {
	log.Printf("Product created: ID=%s, Name=%s, Price=%s, EventID=%s, Source=%s",
		pe.Product.Id,
		pe.Product.Name,
		pe.Product.Price,
		pe.EventId,
		pe.Data.Source,
	)

	// Here you could:
	// - Update search index
	// - Sync with inventory system
	// - Update analytics
	// - Send notifications
	// - Access metadata: pe.Data.Metadata

	return nil
}

func HandleProductUpdated(pe *eventv1.ProductUpdatedEvent, m *message.Message) error {
	log.Printf("Product updated: ID=%s, Name=%s, Price=%s, EventID=%s, Source=%s, ChangedFields=%v",
		pe.Product.Id,
		pe.Product.Name,
		pe.Product.Price,
		pe.EventId,
		pe.Data.Source,
		pe.Data.ChangedFields,
	)

	// Here you could:
	// - Update cached data
	// - Sync with external systems
	// - Update search indexes
	// - Access previous product: pe.Data.PreviousProduct
	// - Access metadata: pe.Data.Metadata

	return nil
}

func HandleProductDeleted(pe *eventv1.ProductDeletedEvent, m *message.Message) error {
	log.Printf("Product deleted: ID=%s, Name=%s, EventID=%s, Source=%s, Reason=%s",
		pe.Product.Id,
		pe.Product.Name,
		pe.EventId,
		pe.Data.Source,
		pe.Data.Reason,
	)

	// Here you could:
	// - Remove from search index
	// - Clean up related data
	// - Update analytics
	// - Archive product information
	// - Access metadata: pe.Data.Metadata

	return nil
}

func HandleProductPriceChanged(pe *eventv1.ProductPriceChangedEvent, m *message.Message) error {
	log.Printf("Product price changed: ID=%s, Name=%s, PreviousPrice=%s, NewPrice=%s, EventID=%s, Source=%s",
		pe.Product.Id,
		pe.Product.Name,
		pe.Data.PreviousPrice,
		pe.Data.NewPrice,
		pe.EventId,
		pe.Data.Source,
	)

	// Here you could:
	// - Update pricing alerts
	// - Recalculate recommendations
	// - Update analytics dashboards
	// - Send price change notifications
	// - Access metadata: pe.Data.Metadata

	return nil
}