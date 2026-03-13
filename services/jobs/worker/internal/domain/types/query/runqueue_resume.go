package query

// RunQueueCreatePendingResumeParams describes one idempotent pending-resume insert derived from an existing run.
type RunQueueCreatePendingResumeParams struct {
	// SourceRunID identifies an existing run whose payload/agent/project should be reused for resume.
	SourceRunID string
	// CorrelationID deduplicates resume scheduling across worker retries and duplicate terminal outcomes.
	CorrelationID string
}
