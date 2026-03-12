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
	// RunLeaseTTL defines ownership lease duration for running run reconciliation.
	RunLeaseTTL time.Duration

	// ProjectLearningModeDefault is the default learning-mode flag to apply when auto-creating projects.
	ProjectLearningModeDefault bool
}

// RunQueueClaimRunningParams defines constraints for claiming running runs for reconciliation.
type RunQueueClaimRunningParams struct {
	// WorkerID identifies worker instance claiming running runs.
	WorkerID string
	// LeaseTTL defines reconciliation ownership lease duration.
	LeaseTTL time.Duration
	// Limit caps number of running runs claimed per tick.
	Limit int
}

// RunQueueClaimedRun represents a pending run promoted into running state.
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
	// SlotNo is a slot number leased for this run, zero when lease is not required.
	SlotNo int
	// SlotID is a unique slot identifier, empty when lease is not required.
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
	// LeaseOwner identifies reconciler that currently owns run lease.
	LeaseOwner string
	// Status must be succeeded, failed, or canceled.
	Status rundomain.Status
	// FinishedAt is a final status timestamp.
	FinishedAt time.Time
}

// RunQueueExtendLeaseParams describes slot lease keepalive update for one running run.
type RunQueueExtendLeaseParams struct {
	// RunID is a running run that currently owns slot lease.
	RunID string
	// ProjectID scopes slot ownership lookup.
	ProjectID string
	// LeaseTTL extends lease_until from current timestamp.
	LeaseTTL time.Duration
}
