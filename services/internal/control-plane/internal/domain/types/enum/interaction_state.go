package enum

// InteractionState describes coarse aggregate lifecycle state.
type InteractionState string

const (
	InteractionStatePendingDispatch   InteractionState = "pending_dispatch"
	InteractionStateOpen              InteractionState = "open"
	InteractionStateResolved          InteractionState = "resolved"
	InteractionStateExpired           InteractionState = "expired"
	InteractionStateDeliveryExhausted InteractionState = "delivery_exhausted"
	InteractionStateCancelled         InteractionState = "cancelled"
)
