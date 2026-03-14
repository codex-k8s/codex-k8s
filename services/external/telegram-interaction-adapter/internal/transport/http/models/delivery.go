package models

import "time"

// TelegramInteractionDeliveryEnvelope is the typed HTTP DTO accepted from worker.
type TelegramInteractionDeliveryEnvelope struct {
	SchemaVersion      string                           `json:"schema_version"`
	DeliveryID         string                           `json:"delivery_id"`
	DeliveryRole       string                           `json:"delivery_role"`
	InteractionID      string                           `json:"interaction_id"`
	InteractionKind    string                           `json:"interaction_kind"`
	RecipientProvider  string                           `json:"recipient_provider"`
	RecipientRef       string                           `json:"recipient_ref"`
	Locale             string                           `json:"locale,omitempty"`
	ContextLinks       InteractionContextLinks          `json:"context_links"`
	Content            *TelegramInteractionContent      `json:"content,omitempty"`
	CallbackEndpoint   *TelegramCallbackEndpoint        `json:"callback_endpoint,omitempty"`
	ProviderMessageRef *TelegramProviderMessageRef      `json:"provider_message_ref,omitempty"`
	Continuation       *TelegramInteractionContinuation `json:"continuation,omitempty"`
	ContinuationPolicy TelegramContinuationPolicy       `json:"continuation_policy"`
	DeliveryDeadlineAt *time.Time                       `json:"delivery_deadline_at,omitempty"`
}

// InteractionContextLinks keeps platform deep links and correlation references.
type InteractionContextLinks struct {
	RunID              string  `json:"run_id"`
	RunURL             *string `json:"run_url,omitempty"`
	IssueURL           *string `json:"issue_url,omitempty"`
	PullRequestURL     *string `json:"pull_request_url,omitempty"`
	RepositoryFullName *string `json:"repository_full_name,omitempty"`
}

// TelegramInteractionContent keeps notify/decision content in one closed DTO.
type TelegramInteractionContent struct {
	NotificationKind    *string                  `json:"notification_kind,omitempty"`
	Summary             *string                  `json:"summary,omitempty"`
	DetailsMarkdown     *string                  `json:"details_markdown,omitempty"`
	ActionLabel         *string                  `json:"action_label,omitempty"`
	ActionURL           *string                  `json:"action_url,omitempty"`
	Question            *string                  `json:"question,omitempty"`
	Options             []TelegramDecisionOption `json:"options,omitempty"`
	AllowFreeText       *bool                    `json:"allow_free_text,omitempty"`
	FreeTextPlaceholder *string                  `json:"free_text_placeholder,omitempty"`
	ExpiresAt           *time.Time               `json:"expires_at,omitempty"`
	ReplyInstruction    *string                  `json:"reply_instruction,omitempty"`
}

// TelegramDecisionOption describes one decision button.
type TelegramDecisionOption struct {
	OptionID       string  `json:"option_id"`
	Label          string  `json:"label"`
	Description    *string `json:"description,omitempty"`
	CallbackHandle string  `json:"callback_handle"`
}

// TelegramCallbackEndpoint contains normalized callback configuration.
type TelegramCallbackEndpoint struct {
	URL            string                   `json:"url"`
	BearerToken    string                   `json:"bearer_token"`
	TokenExpiresAt *time.Time               `json:"token_expires_at,omitempty"`
	Handles        []TelegramCallbackHandle `json:"handles"`
}

// TelegramCallbackHandle describes one opaque callback handle.
type TelegramCallbackHandle struct {
	Handle      string     `json:"handle"`
	HandleKind  string     `json:"handle_kind"`
	ButtonLabel *string    `json:"button_label,omitempty"`
	OptionID    *string    `json:"option_id,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// TelegramProviderMessageRef stores provider-side message identifiers.
type TelegramProviderMessageRef struct {
	ChatRef         *string    `json:"chat_ref,omitempty"`
	MessageID       *string    `json:"message_id,omitempty"`
	InlineMessageID *string    `json:"inline_message_id,omitempty"`
	SentAt          *time.Time `json:"sent_at,omitempty"`
}

// TelegramInteractionContinuation keeps continuation metadata selected by platform.
type TelegramInteractionContinuation struct {
	Action         string     `json:"action"`
	Reason         *string    `json:"reason,omitempty"`
	ResolutionKind *string    `json:"resolution_kind,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

// TelegramContinuationPolicy keeps continuation toggles selected by platform.
type TelegramContinuationPolicy struct {
	PreferredMode                   string `json:"preferred_mode"`
	DisableKeyboardOnResolution     bool   `json:"disable_keyboard_on_resolution"`
	SendFollowUpOnEditFailure       bool   `json:"send_follow_up_on_edit_failure"`
	ManualFallbackOnFollowUpFailure bool   `json:"manual_fallback_on_follow_up_failure"`
}

// TelegramInteractionDeliveryResponse is the typed worker-facing HTTP response.
type TelegramInteractionDeliveryResponse struct {
	Accepted           bool                        `json:"accepted"`
	AdapterDeliveryID  *string                     `json:"adapter_delivery_id,omitempty"`
	ProviderMessageRef *TelegramProviderMessageRef `json:"provider_message_ref,omitempty"`
	EditCapability     *string                     `json:"edit_capability,omitempty"`
	Retryable          bool                        `json:"retryable"`
	Message            *string                     `json:"message,omitempty"`
}
