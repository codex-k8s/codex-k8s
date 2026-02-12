package query

import "encoding/json"

// AgentRunCreateParams defines data required to create a pending agent run.
type AgentRunCreateParams struct {
	// CorrelationID deduplicates webhook processing across retries.
	CorrelationID string
	// ProjectID optionally assigns run to a configured project scope.
	// When empty, project will be derived by the worker.
	ProjectID string
	// AgentID assigns run to resolved agent profile.
	AgentID string
	// RunPayload stores normalized webhook payload for further processing.
	RunPayload json.RawMessage
	// LearningMode is an effective run-level learning mode flag.
	LearningMode bool
}

// AgentRunCreateResult describes the outcome of idempotent run creation.
type AgentRunCreateResult struct {
	// RunID is either newly created or existing run identifier.
	RunID string
	// Inserted is true when a new run was inserted.
	Inserted bool
}
