package enum

// InteractionCallbackKind describes normalized callback event family.
type InteractionCallbackKind string

const (
	InteractionCallbackKindDeliveryReceipt  InteractionCallbackKind = "delivery_receipt"
	InteractionCallbackKindOptionSelected   InteractionCallbackKind = "option_selected"
	InteractionCallbackKindFreeTextReceived InteractionCallbackKind = "free_text_received"
	InteractionCallbackKindTransportFailure InteractionCallbackKind = "transport_failure"
)
