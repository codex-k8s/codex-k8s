package enum

// InteractionCallbackRecordClassification is persisted callback/response evidence outcome.
type InteractionCallbackRecordClassification string

const (
	InteractionCallbackRecordClassificationApplied   InteractionCallbackRecordClassification = "applied"
	InteractionCallbackRecordClassificationDuplicate InteractionCallbackRecordClassification = "duplicate"
	InteractionCallbackRecordClassificationStale     InteractionCallbackRecordClassification = "stale"
	InteractionCallbackRecordClassificationExpired   InteractionCallbackRecordClassification = "expired"
	InteractionCallbackRecordClassificationInvalid   InteractionCallbackRecordClassification = "invalid"
)

// InteractionCallbackResultClassification is external callback processing outcome.
type InteractionCallbackResultClassification string

const (
	InteractionCallbackResultClassificationAccepted  InteractionCallbackResultClassification = "accepted"
	InteractionCallbackResultClassificationDuplicate InteractionCallbackResultClassification = "duplicate"
	InteractionCallbackResultClassificationStale     InteractionCallbackResultClassification = "stale"
	InteractionCallbackResultClassificationExpired   InteractionCallbackResultClassification = "expired"
	InteractionCallbackResultClassificationInvalid   InteractionCallbackResultClassification = "invalid"
)
