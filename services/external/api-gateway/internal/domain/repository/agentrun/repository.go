package agentrun

import (
	"context"
	"encoding/json"
)

// CreateParams defines data required to create a pending agent run.
type CreateParams struct {
	// CorrelationID deduplicates webhook processing across retries.
	CorrelationID string
	// RunPayload stores normalized webhook payload for further processing.
	RunPayload json.RawMessage
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
}
