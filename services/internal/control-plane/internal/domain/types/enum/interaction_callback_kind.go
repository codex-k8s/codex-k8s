package enum

// InteractionCallbackKind describes normalized callback event family.
type InteractionCallbackKind string

const (
	InteractionCallbackKindDeliveryReceipt  InteractionCallbackKind = "delivery_receipt"
	InteractionCallbackKindDecisionResponse InteractionCallbackKind = "decision_response"
)
