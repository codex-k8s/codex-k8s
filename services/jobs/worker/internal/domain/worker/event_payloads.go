package worker

import (
	"encoding/json"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
)

const payloadMarshalFailedError = "payload_marshal_failed"

// namespaceCleanupSkipReason keeps legacy reason field for backwards-compatible payload decoding.
type namespaceCleanupSkipReason string

// runFailureReason keeps normalized worker-side failure reasons in flow-event payloads.
type runFailureReason string

const (
	runFailureReasonKubernetesJobFailed    runFailureReason = "kubernetes job failed"
	runFailureReasonKubernetesJobNotFound  runFailureReason = "kubernetes job not found"
	runFailureReasonNamespacePrepareFailed runFailureReason = "namespace_prepare_failed"
	runFailureReasonRuntimeDeployFailed    runFailureReason = "runtime_deploy_failed"
	runFailureReasonRuntimeDeployCanceled  runFailureReason = "runtime_deploy_canceled"
	runFailureReasonMCPTokenIssueFailed    runFailureReason = "mcp_token_issue_failed"
	runFailureReasonAgentContextResolve    runFailureReason = "agent_context_resolve_failed"
	runFailureReasonPreconditionFailed     runFailureReason = "failed_precondition"
)

// runStartedEventPayload defines payload shape for run.started flow events.
type runStartedEventPayload struct {
	RunID                string                  `json:"run_id"`
	ProjectID            string                  `json:"project_id"`
	SlotNo               int                     `json:"slot_no"`
	JobName              string                  `json:"job_name"`
	JobNamespace         string                  `json:"job_namespace"`
	RuntimeMode          agentdomain.RuntimeMode `json:"runtime_mode"`
	RepositoryFullName   string                  `json:"repository_full_name,omitempty"`
	AgentKey             string                  `json:"agent_key,omitempty"`
	IssueNumber          int64                   `json:"issue_number,omitempty"`
	TriggerKind          string                  `json:"trigger_kind,omitempty"`
	TriggerLabel         string                  `json:"trigger_label,omitempty"`
	JobImage             string                  `json:"job_image,omitempty"`
	Model                string                  `json:"model,omitempty"`
	ModelSource          string                  `json:"model_source,omitempty"`
	ReasoningEffort      string                  `json:"reasoning_effort,omitempty"`
	ReasoningSource      string                  `json:"reasoning_source,omitempty"`
	PromptTemplateKind   string                  `json:"prompt_template_kind,omitempty"`
	PromptTemplateSource string                  `json:"prompt_template_source,omitempty"`
	PromptTemplateLocale string                  `json:"prompt_template_locale,omitempty"`
	BaseBranch           string                  `json:"base_branch,omitempty"`
}

// runProfileResolvedEventPayload defines payload shape for run.profile.resolved flow events.
type runProfileResolvedEventPayload struct {
	RunID              string `json:"run_id"`
	ProjectID          string `json:"project_id"`
	RepositoryFullName string `json:"repository_full_name,omitempty"`
	IssueNumber        int64  `json:"issue_number,omitempty"`
	PullRequestNumber  int    `json:"pull_request_number,omitempty"`
	TriggerKind        string `json:"trigger_kind,omitempty"`
	Model              string `json:"model,omitempty"`
	ModelSource        string `json:"model_source,omitempty"`
	ReasoningEffort    string `json:"reasoning_effort,omitempty"`
	ReasoningSource    string `json:"reasoning_source,omitempty"`
}

// runJobImageResolvedEventPayload defines payload shape for run.job.image.resolved flow events.
type runJobImageResolvedEventPayload struct {
	RunID             string `json:"run_id"`
	ProjectID         string `json:"project_id"`
	SelectedImage     string `json:"selected_image,omitempty"`
	SelectedSource    string `json:"selected_source,omitempty"`
	PrimaryImage      string `json:"primary_image,omitempty"`
	FallbackImage     string `json:"fallback_image,omitempty"`
	PrimaryAvailable  bool   `json:"primary_available,omitempty"`
	FallbackAvailable bool   `json:"fallback_available,omitempty"`
	CheckError        string `json:"check_error,omitempty"`
}

// runFinishedEventPayload defines payload shape for run finish flow events.
type runFinishedEventPayload struct {
	RunID        string                  `json:"run_id"`
	ProjectID    string                  `json:"project_id"`
	Status       rundomain.Status        `json:"status"`
	JobName      string                  `json:"job_name"`
	JobNamespace string                  `json:"job_namespace"`
	RuntimeMode  agentdomain.RuntimeMode `json:"runtime_mode"`
	Namespace    string                  `json:"namespace,omitempty"`
	Error        string                  `json:"error,omitempty"`
	Reason       runFailureReason        `json:"reason,omitempty"`
}

// runFinishedEventExtra carries optional failure details for run finish payloads.
type runFinishedEventExtra struct {
	Error  string
	Reason runFailureReason
}

// namespaceLifecycleEventPayload defines payload shape for namespace lifecycle flow events.
type namespaceLifecycleEventPayload struct {
	RunID                   string                     `json:"run_id"`
	ProjectID               string                     `json:"project_id"`
	RuntimeMode             agentdomain.RuntimeMode    `json:"runtime_mode"`
	Namespace               string                     `json:"namespace"`
	Error                   string                     `json:"error,omitempty"`
	Reason                  namespaceCleanupSkipReason `json:"reason,omitempty"`
	CleanupCommand          string                     `json:"cleanup_command,omitempty"`
	NamespaceLeaseTTL       string                     `json:"namespace_lease_ttl,omitempty"`
	NamespaceLeaseExpiresAt string                     `json:"namespace_lease_expires_at,omitempty"`
	NamespaceReused         bool                       `json:"namespace_reused,omitempty"`
}

// namespaceLifecycleEventExtra carries optional namespace lifecycle diagnostics.
type namespaceLifecycleEventExtra struct {
	Error                   string
	Reason                  namespaceCleanupSkipReason
	CleanupCommand          string
	NamespaceLeaseTTL       time.Duration
	NamespaceLeaseExpiresAt time.Time
	NamespaceReused         bool
}

// payloadMarshalError is fallback payload shape used when JSON serialization unexpectedly fails.
type payloadMarshalError struct {
	Error string `json:"error"`
}

// encodeRunStartedEventPayload serializes run.started payload with safe fallback JSON.
func encodeRunStartedEventPayload(payload runStartedEventPayload) json.RawMessage {
	bytes, err := json.Marshal(payload)
	return marshalPayload(bytes, err)
}

// encodeRunProfileResolvedEventPayload serializes run.profile.resolved payload with safe fallback JSON.
func encodeRunProfileResolvedEventPayload(payload runProfileResolvedEventPayload) json.RawMessage {
	bytes, err := json.Marshal(payload)
	return marshalPayload(bytes, err)
}

// encodeRunJobImageResolvedEventPayload serializes run.job.image.resolved payload with safe fallback JSON.
func encodeRunJobImageResolvedEventPayload(payload runJobImageResolvedEventPayload) json.RawMessage {
	bytes, err := json.Marshal(payload)
	return marshalPayload(bytes, err)
}

// encodeRunFinishedEventPayload serializes run finish payload with safe fallback JSON.
func encodeRunFinishedEventPayload(payload runFinishedEventPayload) json.RawMessage {
	bytes, err := json.Marshal(payload)
	return marshalPayload(bytes, err)
}

// encodeNamespaceLifecycleEventPayload serializes namespace lifecycle payload with safe fallback JSON.
func encodeNamespaceLifecycleEventPayload(payload namespaceLifecycleEventPayload) json.RawMessage {
	bytes, err := json.Marshal(payload)
	return marshalPayload(bytes, err)
}

// marshalPayload centralizes safe JSON fallback to keep event publishing non-blocking on marshal errors.
func marshalPayload(bytes []byte, err error) json.RawMessage {
	if err == nil {
		return json.RawMessage(bytes)
	}
	fallback, fallbackErr := json.Marshal(payloadMarshalError{Error: payloadMarshalFailedError})
	if fallbackErr != nil {
		return json.RawMessage(`{"error":"payload_marshal_failed"}`)
	}
	return json.RawMessage(fallback)
}
