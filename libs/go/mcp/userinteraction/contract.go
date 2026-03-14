package userinteraction

const (
	// ResumePayloadRunPayloadFieldName stores deterministic interaction resume data inside run payload JSON.
	ResumePayloadRunPayloadFieldName = "interaction_resume_payload"
	// DecisionResponseFreeTextMaxBytes bounds user-provided free_text accepted for deterministic resume handoff.
	DecisionResponseFreeTextMaxBytes = 8 * 1024
	// ResumePayloadMaxBytes bounds the serialized interaction resume payload fetched by agent-runner.
	ResumePayloadMaxBytes = 12 * 1024
)
