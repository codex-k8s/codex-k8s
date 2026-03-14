package enum

// InteractionContinuationState describes the next continuation step for the active Telegram binding.
type InteractionContinuationState string

const (
	InteractionContinuationStatePendingPrimaryDelivery InteractionContinuationState = "pending_primary_delivery"
	InteractionContinuationStateReadyForEdit           InteractionContinuationState = "ready_for_edit"
	InteractionContinuationStateFollowUpRequired       InteractionContinuationState = "follow_up_required"
	InteractionContinuationStateManualFallbackRequired InteractionContinuationState = "manual_fallback_required"
	InteractionContinuationStateClosed                 InteractionContinuationState = "closed"
)
