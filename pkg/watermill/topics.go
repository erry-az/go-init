package watermill

// Topic constants for the application
const (
	// User-related topics
	TopicUserCreated = "user.created"
	TopicUserUpdated = "user.updated"
	TopicUserDeleted = "user.deleted"

	// Product-related topics
	TopicProductCreated     = "product.created"
	TopicProductUpdated     = "product.updated"
	TopicProductDeleted     = "product.deleted"
	TopicProductPriceChanged = "product.price.changed"

	// Analytics topics
	TopicUserAnalytics    = "analytics.user"
	TopicProductAnalytics = "analytics.product"

	// Audit topics
	TopicAuditLog = "audit.log"
)

// GetTopicsToInitialize returns all topics that should be initialized at startup
func GetTopicsToInitialize() []string {
	return []string{
		TopicUserCreated,
		TopicUserUpdated,
		TopicUserDeleted,
		TopicProductCreated,
		TopicProductUpdated,
		TopicProductDeleted,
		TopicProductPriceChanged,
		TopicUserAnalytics,
		TopicProductAnalytics,
		TopicAuditLog,
	}
}