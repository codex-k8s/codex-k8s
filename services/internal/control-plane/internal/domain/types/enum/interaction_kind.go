package enum

// InteractionKind identifies one built-in user interaction family.
type InteractionKind string

const (
	InteractionKindNotify          InteractionKind = "notify"
	InteractionKindDecisionRequest InteractionKind = "decision_request"
)
