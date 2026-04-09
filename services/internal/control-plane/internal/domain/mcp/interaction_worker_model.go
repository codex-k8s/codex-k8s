package mcp

import (
	"encoding/json"
	"time"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

// ClaimNextInteractionDispatchParams describes one worker poll for the next due dispatch attempt.
type ClaimNextInteractionDispatchParams struct {
	PendingAttemptTimeout time.Duration
}

// InteractionDispatchClaim carries one claimed interaction attempt and opaque delivery envelope.
type InteractionDispatchClaim struct {
	CorrelationID       string
	Interaction         entitytypes.InteractionRequest
	Attempt             entitytypes.InteractionDeliveryAttempt
	RequestEnvelopeJSON json.RawMessage
}

// CompleteInteractionDispatchParams describes one persisted dispatch outcome from worker.
type CompleteInteractionDispatchParams = querytypes.InteractionDispatchCompleteParams

// CompleteInteractionDispatchResult describes aggregate state after attempt completion.
type CompleteInteractionDispatchResult struct {
	InteractionID       string
	RunID               string
	InteractionState    enumtypes.InteractionState
	ResumeRequired      bool
	ResumeCorrelationID string
}

// ExpireNextInteractionResult describes one processed due-expiry interaction.
type ExpireNextInteractionResult struct {
	InteractionID       string
	RunID               string
	InteractionState    enumtypes.InteractionState
	ResumeRequired      bool
	ResumeCorrelationID string
}
