package enum

// InteractionDeliveryAttemptStatus describes outbound dispatch attempt state.
type InteractionDeliveryAttemptStatus string

const (
	InteractionDeliveryAttemptStatusPending   InteractionDeliveryAttemptStatus = "pending"
	InteractionDeliveryAttemptStatusAccepted  InteractionDeliveryAttemptStatus = "accepted"
	InteractionDeliveryAttemptStatusDelivered InteractionDeliveryAttemptStatus = "delivered"
	InteractionDeliveryAttemptStatusFailed    InteractionDeliveryAttemptStatus = "failed"
	InteractionDeliveryAttemptStatusExhausted InteractionDeliveryAttemptStatus = "exhausted"
)
