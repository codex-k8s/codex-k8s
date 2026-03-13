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
	InteractionID       string
	AdapterKind         string
	RequestEnvelopeJSON json.RawMessage
	AckPayloadJSON      json.RawMessage
	AdapterDeliveryID   string
	Retryable           bool
	NextRetryAt         *time.Time
	LastErrorCode       string
	Status              enumtypes.InteractionDeliveryAttemptStatus
	StartedAt           time.Time
	FinishedAt          *time.Time
}

// InteractionCallbackApplyParams defines one normalized callback application request.
type InteractionCallbackApplyParams struct {
	InteractionID         string                            `json:"interaction_id"`
	DeliveryID            string                            `json:"delivery_id,omitempty"`
	AdapterEventID        string                            `json:"adapter_event_id"`
	CallbackKind          enumtypes.InteractionCallbackKind `json:"callback_kind"`
	OccurredAt            time.Time                         `json:"occurred_at"`
	DeliveryStatus        string                            `json:"delivery_status,omitempty"`
	ResponseKind          enumtypes.InteractionResponseKind `json:"response_kind,omitempty"`
	SelectedOptionID      string                            `json:"selected_option_id,omitempty"`
	FreeText              string                            `json:"free_text,omitempty"`
	ResponderRef          string                            `json:"responder_ref,omitempty"`
	NormalizedPayloadJSON json.RawMessage                   `json:"normalized_payload_json,omitempty"`
	RawPayloadJSON        json.RawMessage                   `json:"raw_payload_json,omitempty"`
}
