package query

import (
	"encoding/json"
	"time"

	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
)

// RunQueueClaimParams defines constraints for claiming a pending run.
type RunQueueClaimParams struct {
	// WorkerID identifies worker instance that claims and leases slots.
	WorkerID string
	// SlotsPerProject is a slot pool size to ensure per project.
	SlotsPerProject int
	// LeaseTTL defines slot lease duration.
	LeaseTTL time.Duration

	// ProjectLearningModeDefault is the default learning-mode flag to apply when auto-creating projects.
	ProjectLearningModeDefault bool
}

// RunQueueClaimedRun represents a pending run promoted into running state with slot lease.
type RunQueueClaimedRun struct {
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

// RunQueueRunningRun is an active run tracked for reconciliation.
type RunQueueRunningRun struct {
	// RunID is a unique run identifier.
	RunID string
	// CorrelationID links run to webhook flow.
	CorrelationID string
	// ProjectID is an effective project scope.
	ProjectID string
	// SlotID is leased slot identifier when available.
	SlotID string
	// SlotNo is leased slot number when available.
	SlotNo int
	// LearningMode is an effective run learning mode flag.
	LearningMode bool
	// RunPayload stores normalized webhook payload.
	RunPayload json.RawMessage
	// StartedAt is timestamp when run entered running state.
	StartedAt time.Time
}

// RunQueueFinishParams describes final run transition and slot release.
type RunQueueFinishParams struct {
	// RunID is a run to finalize.
	RunID string
	// ProjectID is a project scope used for slot release.
	ProjectID string
	// Status must be succeeded, failed, or canceled.
	Status rundomain.Status
	// FinishedAt is a final status timestamp.
	FinishedAt time.Time
}
