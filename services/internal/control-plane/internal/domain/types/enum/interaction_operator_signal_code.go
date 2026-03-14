package enum

// InteractionOperatorSignalCode identifies the latest operator-visible signal for one interaction.
type InteractionOperatorSignalCode string

const (
	InteractionOperatorSignalCodeNone                  InteractionOperatorSignalCode = ""
	InteractionOperatorSignalCodeDeliveryRetryExhausted InteractionOperatorSignalCode = "delivery_retry_exhausted"
	InteractionOperatorSignalCodeInvalidCallbackPayload InteractionOperatorSignalCode = "invalid_callback_payload"
	InteractionOperatorSignalCodeExpiredWait            InteractionOperatorSignalCode = "expired_wait"
	InteractionOperatorSignalCodeEditFallbackSent       InteractionOperatorSignalCode = "edit_fallback_sent"
	InteractionOperatorSignalCodeFollowUpFailed         InteractionOperatorSignalCode = "follow_up_failed"
	InteractionOperatorSignalCodeManualResumeRequired   InteractionOperatorSignalCode = "manual_resume_required"
)
