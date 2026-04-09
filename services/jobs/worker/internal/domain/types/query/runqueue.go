package query

import (
	"encoding/json"
	"time"

	rundomain "github.com/codex-k8s/kodex/libs/go/domain/run"
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

// RunQueueReleaseStaleLeasesParams describes one stale running-lease sweep.
type RunQueueReleaseStaleLeasesParams struct {
	// Limit caps number of stale leases released in one sweep.
	Limit int
	// ReleaseMissingOwners enables immediate reclaim for leases whose owner has no worker_instances row
	// when Kubernetes confirmed the live worker pod set for the current sweep.
	ReleaseMissingOwners bool
	// ActiveWorkerIDs contains worker ids currently backed by live Kubernetes worker pods.
	ActiveWorkerIDs []string
}

// RunQueueReleaseOwnedLeasesParams describes one graceful running-lease release during worker shutdown.
type RunQueueReleaseOwnedLeasesParams struct {
	// WorkerID identifies worker instance that currently owns running-run leases.
	WorkerID string
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
	// ReclaimedAfterStaleLease indicates the run lease was previously released by stale-worker recovery.
	ReclaimedAfterStaleLease bool
}

// RunQueueReleasedStaleLease describes one run lease released after stale-worker detection.
type RunQueueReleasedStaleLease struct {
	// RunID is a unique run identifier.
	RunID string
	// CorrelationID links run to webhook flow.
	CorrelationID string
	// ProjectID is an effective project scope.
	ProjectID string
	// PreviousLeaseOwner is a worker instance that lost liveness.
	PreviousLeaseOwner string
	// PreviousLeaseUntil is the last run lease deadline held by stale owner.
	PreviousLeaseUntil *time.Time
	// WorkerHeartbeatAt is the last recorded worker heartbeat when available.
	WorkerHeartbeatAt *time.Time
	// WorkerExpiresAt is the worker liveness deadline when available.
	WorkerExpiresAt *time.Time
	// WorkerStatus is the stale worker lifecycle state or "missing" fallback.
	WorkerStatus string
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

// RunQueueNonTerminalRun describes one run that must block namespace cleanup.
type RunQueueNonTerminalRun struct {
	// RunID is a run referenced by managed namespace labels.
	RunID string
	// Status is current non-terminal lifecycle status from agent_runs.
	Status string
}
