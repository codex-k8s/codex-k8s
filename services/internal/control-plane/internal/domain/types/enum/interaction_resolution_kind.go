package enum

// InteractionResolutionKind describes semantic outcome of terminal interaction state.
type InteractionResolutionKind string

const (
	InteractionResolutionKindNone              InteractionResolutionKind = "none"
	InteractionResolutionKindDeliveryOnly      InteractionResolutionKind = "delivery_only"
	InteractionResolutionKindOptionSelected    InteractionResolutionKind = "option_selected"
	InteractionResolutionKindFreeTextSubmitted InteractionResolutionKind = "free_text_submitted"
)
