package service

import "time"

const (
	SchemaVersionTelegramInteractionV1 = "telegram-interaction-v1"

	DeliveryRolePrimaryDispatch = "primary_dispatch"
	DeliveryRoleMessageEdit     = "message_edit"
	DeliveryRoleFollowUpNotify  = "follow_up_notify"

	InteractionKindNotify          = "notify"
	InteractionKindDecisionRequest = "decision_request"

	CallbackKindOptionSelected   = "option_selected"
	CallbackKindFreeTextReceived = "free_text_received"

	ContinuationActionEditMessage  = "edit_message"
	ContinuationActionSendFollowUp = "send_follow_up"

	HandleKindOption          = "option"
	HandleKindFreeTextSession = "free_text_session"

	EditCapabilityUnknown      = "unknown"
	EditCapabilityKeyboardOnly = "keyboard_only"
	EditCapabilityFollowUpOnly = "follow_up_only"
)

const (
	recipientRefPrefixGitHubLogin = "github_login:"
	recipientRefPrefixChatID      = "telegram_chat_id:"
)

// DeliveryEnvelope is the adapter-side input contract derived from worker dispatch payload.
type DeliveryEnvelope struct {
	SchemaVersion      string
	DeliveryID         string
	DeliveryRole       string
	InteractionID      string
	InteractionKind    string
	RecipientProvider  string
	RecipientRef       string
	Locale             string
	ContextLinks       ContextLinks
	Content            InteractionContent
	CallbackEndpoint   *CallbackEndpoint
	ProviderMessageRef *ProviderMessageRef
	Continuation       *Continuation
	ContinuationPolicy ContinuationPolicy
	DeliveryDeadlineAt *time.Time
}

// ContextLinks keeps deep-links from platform delivery envelope.
type ContextLinks struct {
	RunID              string
	RunURL             string
	IssueURL           string
	PullRequestURL     string
	RepositoryFullName string
}

// InteractionContent keeps notify and decision fields in one typed structure.
type InteractionContent struct {
	NotificationKind    string
	Summary             string
	DetailsMarkdown     string
	ActionLabel         string
	ActionURL           string
	Question            string
	Options             []DecisionOption
	AllowFreeText       bool
	FreeTextPlaceholder string
	ExpiresAt           *time.Time
	ReplyInstruction    string
}

// DecisionOption describes one Telegram inline button.
type DecisionOption struct {
	OptionID       string
	Label          string
	Description    string
	CallbackHandle string
}

// CallbackEndpoint describes where adapter must forward normalized callbacks.
type CallbackEndpoint struct {
	URL            string
	BearerToken    string
	TokenExpiresAt *time.Time
	Handles        []CallbackHandle
}

// CallbackHandle describes one opaque callback or free-text session handle.
type CallbackHandle struct {
	Handle      string
	HandleKind  string
	ButtonLabel string
	OptionID    string
	ExpiresAt   time.Time
}

// ContinuationPolicy mirrors platform continuation toggles for edit/follow-up flow.
type ContinuationPolicy struct {
	PreferredMode                   string
	DisableKeyboardOnResolution     bool
	SendFollowUpOnEditFailure       bool
	ManualFallbackOnFollowUpFailure bool
}

// Continuation stores already classified continuation action.
type Continuation struct {
	Action         string
	Reason         string
	ResolutionKind string
	ResolvedAt     *time.Time
}

// ProviderMessageRef stores Telegram message identifiers.
type ProviderMessageRef struct {
	ChatRef         string
	MessageID       string
	InlineMessageID string
	SentAt          *time.Time
}

// DeliveryResponse is the typed HTTP response returned to worker.
type DeliveryResponse struct {
	Accepted           bool
	AdapterDeliveryID  string
	ProviderMessageRef *ProviderMessageRef
	EditCapability     string
	Retryable          bool
	Message            string
}

// CallbackOutcome mirrors api-gateway callback outcome.
type CallbackOutcome struct {
	Accepted           bool   `json:"accepted"`
	Classification     string `json:"classification"`
	InteractionState   string `json:"interaction_state"`
	ResumeRequired     bool   `json:"resume_required"`
	ContinuationAction string `json:"continuation_action"`
	Message            string `json:"message"`
}

// CallbackEnvelope is adapter -> api-gateway normalized callback payload.
type CallbackEnvelope struct {
	SchemaVersion           string              `json:"schema_version"`
	InteractionID           string              `json:"interaction_id"`
	DeliveryID              string              `json:"delivery_id,omitempty"`
	AdapterEventID          string              `json:"adapter_event_id"`
	CallbackKind            string              `json:"callback_kind"`
	OccurredAt              string              `json:"occurred_at"`
	CallbackHandle          string              `json:"callback_handle,omitempty"`
	FreeText                string              `json:"free_text,omitempty"`
	ResponderRef            string              `json:"responder_ref,omitempty"`
	ProviderMessageRef      *ProviderMessageRef `json:"provider_message_ref,omitempty"`
	ProviderUpdateID        string              `json:"provider_update_id,omitempty"`
	ProviderCallbackQueryID string              `json:"provider_callback_query_id,omitempty"`
}

type followUpMessageData struct {
	RunURL         string
	IssueURL       string
	PullRequestURL string
}

// DeliveryError describes a typed adapter rejection that still returns JSON body.
type DeliveryError struct {
	StatusCode int
	Response   DeliveryResponse
}

func (e *DeliveryError) Error() string {
	return e.Response.Message
}
