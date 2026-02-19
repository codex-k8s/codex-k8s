package query

import webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"

// RunRuntimePayload keeps only fields that influence worker runtime decisions.
type RunRuntimePayload struct {
	Project    *RunRuntimeProject    `json:"project"`
	Repository *RunRuntimeRepository `json:"repository"`
	Trigger    *RunRuntimeTrigger    `json:"trigger"`
	Issue      *RunRuntimeIssue      `json:"issue"`
	Runtime    *RunRuntimeProfile    `json:"runtime"`
}

// RunRuntimeProject captures project metadata used by runtime deploy orchestration.
type RunRuntimeProject struct {
	ServicesYAML string `json:"services_yaml"`
}

// RunRuntimeRepository captures repository metadata used by runtime deploy orchestration.
type RunRuntimeRepository struct {
	FullName string `json:"full_name"`
}

// RunRuntimeTrigger captures normalized trigger kind from webhook payload.
type RunRuntimeTrigger struct {
	Kind webhookdomain.TriggerKind `json:"kind"`
}

// RunRuntimeIssue captures optional issue metadata used in namespace naming.
type RunRuntimeIssue struct {
	Number int64 `json:"number"`
}

// RunRuntimeProfile captures runtime mode resolved upstream by webhook/control-plane.
type RunRuntimeProfile struct {
	Mode       string `json:"mode"`
	TargetEnv  string `json:"target_env,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	BuildRef   string `json:"build_ref,omitempty"`
	DeployOnly bool   `json:"deploy_only,omitempty"`
}

// RepositoryPayload keeps repository fields required for project derivation.
type RepositoryPayload struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
}

// RunQueuePayload keeps only payload fields required in runqueue repository.
type RunQueuePayload struct {
	Repository RepositoryPayload  `json:"repository"`
	Runtime    *RunRuntimeProfile `json:"runtime,omitempty"`
}

// ProjectSettings stores project-level defaults in JSONB settings.
type ProjectSettings struct {
	LearningModeDefault bool `json:"learning_mode_default"`
	SlotsPerProject     int  `json:"slots_per_project,omitempty"`
}
