package mcp

import (
	"encoding/json"
	"time"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
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
type CompleteInteractionDispatchParams struct {
	InteractionID       string
	DeliveryID          string
	AdapterKind         string
	Status              enumtypes.InteractionDeliveryAttemptStatus
	RequestEnvelopeJSON json.RawMessage
	AckPayloadJSON      json.RawMessage
	AdapterDeliveryID   string
	Retryable           bool
	NextRetryAt         *time.Time
	LastErrorCode       string
	FinishedAt          time.Time
}

// CompleteInteractionDispatchResult describes aggregate state after attempt completion.
type CompleteInteractionDispatchResult struct {
	InteractionID    string
	InteractionState enumtypes.InteractionState
	ResumeRequired   bool
}

// ExpireNextInteractionResult describes one processed due-expiry interaction.
type ExpireNextInteractionResult struct {
	InteractionID    string
	InteractionState enumtypes.InteractionState
	ResumeRequired   bool
}
