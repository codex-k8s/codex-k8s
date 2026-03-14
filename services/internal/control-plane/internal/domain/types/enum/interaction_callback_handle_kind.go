package enum

// InteractionCallbackHandleKind identifies the semantic target behind one opaque callback handle.
type InteractionCallbackHandleKind string

const (
	InteractionCallbackHandleKindOption          InteractionCallbackHandleKind = "option"
	InteractionCallbackHandleKindFreeTextSession InteractionCallbackHandleKind = "free_text_session"
)
