package models

// MCPApprovalCallbackRequest describes decision callback from external approver/executor adapters.
type MCPApprovalCallbackRequest struct {
	ApprovalRequestID int64  `json:"approval_request_id"`
	Decision          string `json:"decision"`
	Reason            string `json:"reason"`
	ActorID           string `json:"actor_id"`
}

// InteractionCallbackEnvelope describes one user interaction callback from an external adapter.
type InteractionCallbackEnvelope struct {
	SchemaVersion           string                         `json:"schema_version"`
	InteractionID           string                         `json:"interaction_id"`
	DeliveryID              string                         `json:"delivery_id,omitempty"`
	AdapterEventID          string                         `json:"adapter_event_id"`
	CallbackKind            string                         `json:"callback_kind"`
	OccurredAt              string                         `json:"occurred_at"`
	CallbackHandle          string                         `json:"callback_handle,omitempty"`
	FreeText                string                         `json:"free_text,omitempty"`
	ResponderRef            string                         `json:"responder_ref,omitempty"`
	ProviderMessageRef      *InteractionProviderMessageRef `json:"provider_message_ref,omitempty"`
	ProviderUpdateID        string                         `json:"provider_update_id,omitempty"`
	ProviderCallbackQueryID string                         `json:"provider_callback_query_id,omitempty"`
	DeliveryStatus          string                         `json:"delivery_status,omitempty"`
	Error                   *InteractionCallbackError      `json:"error,omitempty"`
}

// InteractionProviderMessageRef keeps adapter-side message identifiers opaque but typed.
type InteractionProviderMessageRef struct {
	ChatRef         string `json:"chat_ref,omitempty"`
	MessageID       string `json:"message_id,omitempty"`
	InlineMessageID string `json:"inline_message_id,omitempty"`
	SentAt          string `json:"sent_at,omitempty"`
}

// InteractionCallbackError stores adapter-side failure details that do not alter classification directly.
type InteractionCallbackError struct {
	Code      string `json:"code,omitempty"`
	Retryable bool   `json:"retryable"`
	Message   string `json:"message,omitempty"`
}

// InteractionCallbackOutcome is the typed HTTP response for interaction callbacks.
type InteractionCallbackOutcome struct {
	Accepted           bool   `json:"accepted"`
	Classification     string `json:"classification"`
	InteractionState   string `json:"interaction_state"`
	ResumeRequired     bool   `json:"resume_required"`
	ContinuationAction string `json:"continuation_action,omitempty"`
	Message            string `json:"message,omitempty"`
}
