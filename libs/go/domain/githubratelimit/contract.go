package githubratelimit

const (
	// ResumePayloadRunPayloadFieldName stores deterministic GitHub rate-limit resume data inside run payload JSON.
	ResumePayloadRunPayloadFieldName = "github_rate_limit_resume_payload"
	// ResumeCorrelationPrefix marks pending runs scheduled specifically for deterministic GitHub rate-limit resume.
	ResumeCorrelationPrefix = "github-rate-limit-resume:"
	// RunnerActionPersistSessionAndExitWait tells agent-runner to persist the latest session snapshot and stop local retries.
	RunnerActionPersistSessionAndExitWait = "persist_session_and_exit_wait"
	// ResumePayloadMaxBytes bounds the serialized GitHub rate-limit resume payload fetched by agent-runner.
	ResumePayloadMaxBytes = 12 * 1024
	// SignalExcerptMaxBytes bounds sanitized stderr/message excerpts stored for one GitHub rate-limit signal.
	SignalExcerptMaxBytes = 4 * 1024
)
