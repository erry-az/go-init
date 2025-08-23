package consumer

import (
	"context"
	"log"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
)

type UserConsumer struct{}

func NewUserConsumer() *UserConsumer {
	return &UserConsumer{}
}

func (u *UserConsumer) AddHandlers(eventProcessor *cqrs.EventProcessor) error {
	return eventProcessor.AddHandlers(
		cqrs.NewEventHandler("HandleUserCreated", u.HandleUserCreated),
		cqrs.NewEventHandler("HandleUserUpdated", u.HandleUserUpdated),
		cqrs.NewEventHandler("HandleUserDeleted", u.HandleUserDeleted),
	)
}

func (u *UserConsumer) HandleUserCreated(ctx context.Context, pe *eventv1.UserCreatedEvent) error {
	log.Printf("User created: ID=%s, Name=%s, Email=%s, EventID=%s, Source=%s",
		pe.User.Id,
		pe.User.Name,
		pe.User.Email,
		pe.EventId,
		pe.Data.Source,
	)

	// Here you could:
	// - Send welcome email
	// - Create user profile in another service
	// - Update analytics
	// - Log audit trail
	// - Access metadata: pe.Data.Metadata

	return nil
}

func (u *UserConsumer) HandleUserUpdated(ctx context.Context, pe *eventv1.UserUpdatedEvent) error {
	log.Printf("User updated: ID=%s, Name=%s, Email=%s, EventID=%s, Source=%s, ChangedFields=%v",
		pe.User.Id,
		pe.User.Name,
		pe.User.Email,
		pe.EventId,
		pe.Data.Source,
		pe.Data.ChangedFields,
	)

	// Here you could:
	// - Update cached user data
	// - Sync with external systems
	// - Update search indexes
	// - Access previous user: pe.Data.PreviousUser
	// - Access metadata: pe.Data.Metadata

	return nil
}

func (u *UserConsumer) HandleUserDeleted(ctx context.Context, pe *eventv1.UserDeletedEvent) error {
	log.Printf("User deleted: ID=%s, Name=%s, EventID=%s, Source=%s, Reason=%s",
		pe.User.Id,
		pe.User.Name,
		pe.EventId,
		pe.Data.Source,
		pe.Data.Reason,
	)

	// Here you could:
	// - Clean up user data
	// - Cancel subscriptions
	// - Archive user information
	// - Update analytics
	// - Access metadata: pe.Data.Metadata

	return nil
}
