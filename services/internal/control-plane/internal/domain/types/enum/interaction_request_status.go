package enum

// InteractionRequestStatus describes terminal status exposed to deterministic resume payload.
type InteractionRequestStatus string

const (
	InteractionRequestStatusAnswered          InteractionRequestStatus = "answered"
	InteractionRequestStatusExpired           InteractionRequestStatus = "expired"
	InteractionRequestStatusDeliveryExhausted InteractionRequestStatus = "delivery_exhausted"
	InteractionRequestStatusCancelled         InteractionRequestStatus = "cancelled"
)
