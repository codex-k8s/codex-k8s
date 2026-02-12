package agentrun

import (
	"context"
	"encoding/json"
)

// Run stores the subset of agent_runs required by runtime services.
type Run struct {
	ID            string
	CorrelationID string
	ProjectID     string
	Status        string
	RunPayload    json.RawMessage
}

// CreateParams defines data required to create a pending agent run.
type CreateParams struct {
	// CorrelationID deduplicates webhook processing across retries.
	CorrelationID string
	// ProjectID optionally assigns run to a configured project scope.
	// When empty, project will be derived by the worker.
	ProjectID string
	// RunPayload stores normalized webhook payload for further processing.
	RunPayload json.RawMessage
	// LearningMode is an effective run-level learning mode flag.
	LearningMode bool
}

// CreateResult describes the outcome of idempotent run creation.
type CreateResult struct {
	// RunID is either newly created or existing run identifier.
	RunID string
	// Inserted is true when a new run was inserted.
	Inserted bool
}

// Repository persists and queries agent run records.
type Repository interface {
	// CreatePendingIfAbsent inserts a pending run unless it already exists.
	CreatePendingIfAbsent(ctx context.Context, params CreateParams) (CreateResult, error)
	// GetByID returns one run by id.
	GetByID(ctx context.Context, runID string) (Run, bool, error)
}
