package consumer

import (
	"log"

	"github.com/ThreeDotsLabs/watermill/message"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
)

func HandleUserCreated(pe *eventv1.UserCreatedEvent, m *message.Message) error {
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

func HandleUserUpdated(pe *eventv1.UserUpdatedEvent, m *message.Message) error {
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

func HandleUserDeleted(pe *eventv1.UserDeletedEvent, m *message.Message) error {
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
