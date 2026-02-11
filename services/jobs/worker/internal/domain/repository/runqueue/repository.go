package runqueue

import (
	"context"
	"encoding/json"
	"time"

	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
)

// ClaimParams defines constraints for claiming a pending run.
type ClaimParams struct {
	// WorkerID identifies worker instance that claims and leases slots.
	WorkerID string
	// SlotsPerProject is a slot pool size to ensure per project.
	SlotsPerProject int
	// LeaseTTL defines slot lease duration.
	LeaseTTL time.Duration

	// ProjectLearningModeDefault is the default learning-mode flag to apply when auto-creating projects.
	ProjectLearningModeDefault bool
}

// ClaimedRun represents a pending run promoted into running state with slot lease.
type ClaimedRun struct {
	// RunID is a unique run identifier.
	RunID string
	// CorrelationID links run to webhook flow.
	CorrelationID string
	// ProjectID is an effective project scope used for slot leasing.
	ProjectID string
	// LearningMode is effective run learning mode flag.
	LearningMode bool
	// RunPayload stores normalized webhook payload.
	RunPayload json.RawMessage
	// SlotNo is a slot number leased for this run.
	SlotNo int
	// SlotID is a unique slot identifier.
	SlotID string
}

// RunningRun is an active run tracked for reconciliation.
type RunningRun struct {
	// RunID is a unique run identifier.
	RunID string
	// CorrelationID links run to webhook flow.
	CorrelationID string
	// ProjectID is an effective project scope.
	ProjectID string
	// LearningMode is an effective run learning mode flag.
	LearningMode bool
	// StartedAt is timestamp when run entered running state.
	StartedAt time.Time
}

// FinishParams describes final run transition and slot release.
type FinishParams struct {
	// RunID is a run to finalize.
	RunID string
	// ProjectID is a project scope used for slot release.
	ProjectID string
	// Status must be succeeded, failed, or canceled.
	Status rundomain.Status
	// FinishedAt is a final status timestamp.
	FinishedAt time.Time
}

// Repository provides queue-like operations over agent runs and slots.
type Repository interface {
	// ClaimNextPending atomically claims one pending run and leases a free slot.
	ClaimNextPending(ctx context.Context, params ClaimParams) (ClaimedRun, bool, error)
	// ListRunning returns active runs for reconciliation.
	ListRunning(ctx context.Context, limit int) ([]RunningRun, error)
	// FinishRun finalizes run status and releases slot lease.
	FinishRun(ctx context.Context, params FinishParams) (bool, error)
}
