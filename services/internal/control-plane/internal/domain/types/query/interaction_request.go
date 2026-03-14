package query

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionRequestCreateParams defines one canonical interaction aggregate insert.
type InteractionRequestCreateParams struct {
	ProjectID          string
	RunID              string
	InteractionKind    enumtypes.InteractionKind
	ChannelFamily      enumtypes.InteractionChannelFamily
	State              enumtypes.InteractionState
	ResolutionKind     enumtypes.InteractionResolutionKind
	RecipientProvider  string
	RecipientRef       string
	RequestPayloadJSON json.RawMessage
	ContextLinksJSON   json.RawMessage
	ResponseDeadlineAt *time.Time
}

// InteractionDeliveryAttemptCreateParams defines one dispatch-attempt insert.
type InteractionDeliveryAttemptCreateParams struct {
	InteractionID          string
	ChannelBindingID       int64
	AdapterKind            string
	DeliveryRole           enumtypes.InteractionDeliveryRole
	RequestEnvelopeJSON    json.RawMessage
	AckPayloadJSON         json.RawMessage
	AdapterDeliveryID      string
	ProviderMessageRefJSON json.RawMessage
	Retryable              bool
	NextRetryAt            *time.Time
	LastErrorCode          string
	ContinuationReason     string
	Status                 enumtypes.InteractionDeliveryAttemptStatus
	StartedAt              time.Time
	FinishedAt             *time.Time
}

// InteractionCallbackApplyParams defines one normalized callback application request.
type InteractionCallbackApplyParams struct {
	InteractionID           string                            `json:"interaction_id"`
	DeliveryID              string                            `json:"delivery_id,omitempty"`
	AdapterEventID          string                            `json:"adapter_event_id"`
	CallbackKind            enumtypes.InteractionCallbackKind `json:"callback_kind"`
	OccurredAt              time.Time                         `json:"occurred_at"`
	CallbackHandle          string                            `json:"callback_handle,omitempty"`
	DeliveryStatus          string                            `json:"delivery_status,omitempty"`
	FreeText                string                            `json:"free_text,omitempty"`
	ResponderRef            string                            `json:"responder_ref,omitempty"`
	ProviderMessageRefJSON  json.RawMessage                   `json:"provider_message_ref_json,omitempty"`
	ProviderUpdateID        string                            `json:"provider_update_id,omitempty"`
	ProviderCallbackQueryID string                            `json:"provider_callback_query_id,omitempty"`
	TransportErrorCode      string                            `json:"transport_error_code,omitempty"`
	TransportRetryable      bool                              `json:"transport_retryable,omitempty"`
	NormalizedPayloadJSON   json.RawMessage                   `json:"normalized_payload_json,omitempty"`
	RawPayloadJSON          json.RawMessage                   `json:"raw_payload_json,omitempty"`
}
