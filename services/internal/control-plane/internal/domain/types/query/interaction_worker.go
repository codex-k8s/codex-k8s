package query

import (
	"encoding/json"
	"time"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionDispatchClaimParams describes one worker claim for the next due dispatch attempt.
type InteractionDispatchClaimParams struct {
	Now                   time.Time
	PendingAttemptTimeout time.Duration
}

// InteractionDispatchClaim carries one claimed aggregate and delivery attempt for worker dispatch.
type InteractionDispatchClaim struct {
	Interaction entitytypes.InteractionRequest
	Attempt     entitytypes.InteractionDeliveryAttempt
	Binding     *entitytypes.InteractionChannelBinding
}

// InteractionDispatchCompleteParams describes one persisted dispatch attempt outcome.
type InteractionDispatchCompleteParams struct {
	InteractionID          string
	DeliveryID             string
	AdapterKind            string
	Status                 enumtypes.InteractionDeliveryAttemptStatus
	RequestEnvelopeJSON    json.RawMessage
	AckPayloadJSON         json.RawMessage
	AdapterDeliveryID      string
	ProviderMessageRefJSON json.RawMessage
	EditCapability         enumtypes.InteractionEditCapability
	Retryable              bool
	NextRetryAt            *time.Time
	LastErrorCode          string
	CallbackTokenKeyID     string
	CallbackTokenExpiresAt *time.Time
	FinishedAt             time.Time
}

// InteractionDispatchCompleteResult reports aggregate mutation after dispatch completion.
type InteractionDispatchCompleteResult struct {
	Interaction    entitytypes.InteractionRequest
	Attempt        entitytypes.InteractionDeliveryAttempt
	StateChanged   bool
	ResumeRequired bool
}

// InteractionExpireDueParams describes one worker expiry scan checkpoint.
type InteractionExpireDueParams struct {
	Now time.Time
}

// InteractionExpireDueResult reports aggregate mutation after deadline-based expiry handling.
type InteractionExpireDueResult struct {
	Interaction    entitytypes.InteractionRequest
	Attempt        *entitytypes.InteractionDeliveryAttempt
	StateChanged   bool
	ResumeRequired bool
}
