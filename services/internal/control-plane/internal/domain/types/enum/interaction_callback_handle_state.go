package enum

// InteractionCallbackHandleState describes the lifecycle state of one callback handle.
type InteractionCallbackHandleState string

const (
	InteractionCallbackHandleStateOpen    InteractionCallbackHandleState = "open"
	InteractionCallbackHandleStateUsed    InteractionCallbackHandleState = "used"
	InteractionCallbackHandleStateExpired InteractionCallbackHandleState = "expired"
	InteractionCallbackHandleStateRevoked InteractionCallbackHandleState = "revoked"
)
