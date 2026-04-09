package value

import enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"

type GitHubRateLimitRunStatusCommentRetryTarget struct {
	RunID string `json:"run_id"`
	Phase string `json:"phase"`
}

type GitHubRateLimitRunStatusCommentRetryExecution struct {
	JobName      string `json:"job_name,omitempty"`
	JobNamespace string `json:"job_namespace,omitempty"`
	RuntimeMode  string `json:"runtime_mode,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	TriggerKind  string `json:"trigger_kind,omitempty"`
}

type GitHubRateLimitRunStatusCommentRetryRender struct {
	PromptLocale             string `json:"prompt_locale,omitempty"`
	Model                    string `json:"model,omitempty"`
	ReasoningEffort          string `json:"reasoning_effort,omitempty"`
	RunStatus                string `json:"run_status,omitempty"`
	CodexAuthVerificationURL string `json:"codex_auth_verification_url,omitempty"`
	CodexAuthUserCode        string `json:"codex_auth_user_code,omitempty"`
}

// GitHubRateLimitRunStatusCommentRetryPayload stores deterministic comment retry parameters.
type GitHubRateLimitRunStatusCommentRetryPayload struct {
	GitHubRateLimitRunStatusCommentRetryTarget
	GitHubRateLimitRunStatusCommentRetryExecution
	GitHubRateLimitRunStatusCommentRetryRender
	Deleted        bool `json:"deleted,omitempty"`
	AlreadyDeleted bool `json:"already_deleted,omitempty"`
}

// GitHubRateLimitPlatformCallReplayPayload stores deterministic platform-managed replay parameters.
type GitHubRateLimitPlatformCallReplayPayload struct {
	OperationKind            enumtypes.GitHubRateLimitPlatformReplayOperationKind `json:"operation_kind"`
	RepositoryFullName       string                                               `json:"repository_full_name"`
	IssueNumber              int                                                  `json:"issue_number,omitempty"`
	TargetLabel              string                                               `json:"target_label,omitempty"`
	RequestFingerprint       string                                               `json:"request_fingerprint,omitempty"`
	CorrelationID            string                                               `json:"correlation_id,omitempty"`
	ExpectedCurrentRunLabels []string                                             `json:"expected_current_run_labels,omitempty"`
}
