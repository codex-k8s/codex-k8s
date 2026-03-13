package worker

import (
	"context"
	"encoding/json"
	"time"
)

// InteractionDispatchClaim describes one claimed interaction attempt from control-plane.
type InteractionDispatchClaim struct {
	CorrelationID       string
	InteractionID       string
	RunID               string
	InteractionKind     string
	RecipientProvider   string
	RecipientRef        string
	ResponseDeadlineAt  *time.Time
	Attempt             InteractionDispatchAttempt
	RequestEnvelopeJSON json.RawMessage
}

// InteractionDispatchAttempt identifies one persisted delivery attempt ledger row.
type InteractionDispatchAttempt struct {
	ID          int64
	AttemptNo   int
	DeliveryID  string
	AdapterKind string
}

// CompleteInteractionDispatchParams describes one dispatch completion callback to control-plane.
type CompleteInteractionDispatchParams struct {
	InteractionID       string
	DeliveryID          string
	AdapterKind         string
	Status              string
	RequestEnvelopeJSON json.RawMessage
	AckPayloadJSON      json.RawMessage
	AdapterDeliveryID   string
	Retryable           bool
	NextRetryAt         *time.Time
	LastErrorCode       string
	FinishedAt          time.Time
}

// CompleteInteractionDispatchResult returns aggregate state after dispatch completion.
type CompleteInteractionDispatchResult struct {
	InteractionID       string
	RunID               string
	InteractionState    string
	ResumeRequired      bool
	ResumeCorrelationID string
}

// ExpireNextInteractionResult describes one processed due-expiry interaction.
type ExpireNextInteractionResult struct {
	Found               bool
	InteractionID       string
	RunID               string
	InteractionState    string
	ResumeRequired      bool
	ResumeCorrelationID string
}

// InteractionLifecycleClient exposes worker-side control-plane interaction lifecycle RPCs.
type InteractionLifecycleClient interface {
	ClaimNextInteractionDispatch(ctx context.Context, pendingAttemptTimeout time.Duration) (InteractionDispatchClaim, bool, error)
	CompleteInteractionDispatch(ctx context.Context, params CompleteInteractionDispatchParams) (CompleteInteractionDispatchResult, error)
	ExpireNextInteraction(ctx context.Context) (ExpireNextInteractionResult, error)
}

// InteractionDispatchAck is adapter transport feedback used to classify one attempt.
type InteractionDispatchAck struct {
	AdapterKind       string
	AdapterDeliveryID string
	AckPayloadJSON    json.RawMessage
	Retryable         bool
	ErrorCode         string
}

// InteractionDispatcher sends interaction envelopes to one external adapter family.
type InteractionDispatcher interface {
	Dispatch(ctx context.Context, claim InteractionDispatchClaim) (InteractionDispatchAck, error)
}
