package query

import webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"

// RunRuntimePayload keeps only fields that influence worker runtime decisions.
type RunRuntimePayload struct {
	Trigger *RunRuntimeTrigger `json:"trigger"`
	Issue   *RunRuntimeIssue   `json:"issue"`
}

// RunRuntimeTrigger captures normalized trigger kind from webhook payload.
type RunRuntimeTrigger struct {
	Kind webhookdomain.TriggerKind `json:"kind"`
}

// RunRuntimeIssue captures optional issue metadata used in namespace naming.
type RunRuntimeIssue struct {
	Number int64 `json:"number"`
}

// RepositoryPayload keeps repository fields required for project derivation.
type RepositoryPayload struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
}

// RunQueuePayload keeps only payload fields required in runqueue repository.
type RunQueuePayload struct {
	Repository RepositoryPayload `json:"repository"`
}

// ProjectSettings stores project-level defaults in JSONB settings.
type ProjectSettings struct {
	LearningModeDefault bool `json:"learning_mode_default"`
}
