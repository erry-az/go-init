package consumer

import (
	"context"
	"log"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
)

type ProductConsumer struct{}

func NewProductConsumer() *ProductConsumer {
	return &ProductConsumer{}
}

func (p *ProductConsumer) AddHandlers(eventProcessor *cqrs.EventProcessor) error {
	return eventProcessor.AddHandlers(
		cqrs.NewEventHandler("HandleProductCreated", p.HandleProductCreated),
		cqrs.NewEventHandler("HandleProductUpdated", p.HandleProductUpdated),
		cqrs.NewEventHandler("HandleProductDeleted", p.HandleProductDeleted),
		cqrs.NewEventHandler("HandleProductPriceChanged", p.HandleProductPriceChanged),
	)
}

func (p *ProductConsumer) HandleProductCreated(ctx context.Context, pe *eventv1.ProductCreatedEvent) error {
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

func (p *ProductConsumer) HandleProductUpdated(ctx context.Context, pe *eventv1.ProductUpdatedEvent) error {
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

func (p *ProductConsumer) HandleProductDeleted(ctx context.Context, pe *eventv1.ProductDeletedEvent) error {
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

func (p *ProductConsumer) HandleProductPriceChanged(ctx context.Context, pe *eventv1.ProductPriceChangedEvent) error {
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
