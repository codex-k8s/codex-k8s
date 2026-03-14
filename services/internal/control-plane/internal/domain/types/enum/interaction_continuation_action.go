package enum

// InteractionContinuationAction is the async side effect selected after callback classification.
type InteractionContinuationAction string

const (
	InteractionContinuationActionNone          InteractionContinuationAction = "none"
	InteractionContinuationActionEditMessage   InteractionContinuationAction = "edit_message"
	InteractionContinuationActionSendFollowUp  InteractionContinuationAction = "send_follow_up"
	InteractionContinuationActionManualFallback InteractionContinuationAction = "manual_fallback"
)
