package enum

// InteractionResponseKind describes one typed user answer shape.
type InteractionResponseKind string

const (
	InteractionResponseKindOption   InteractionResponseKind = "option"
	InteractionResponseKindFreeText InteractionResponseKind = "free_text"
	InteractionResponseKindNone     InteractionResponseKind = "none"
)
