package mcp

import (
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

// ToolName is a stable MCP tool identifier.
type ToolName string

const ToolPromptContextGet ToolName = "codex_prompt_context_get"

const (
	ToolGitHubIssueGet           ToolName = "github_issue_get"
	ToolGitHubPullRequestGet     ToolName = "github_pull_request_get"
	ToolGitHubIssueComments      ToolName = "github_issue_comments_list"
	ToolGitHubLabelsList         ToolName = "github_labels_list"
	ToolGitHubBranchesList       ToolName = "github_branches_list"
	ToolGitHubBranchEnsure       ToolName = "github_branch_ensure"
	ToolGitHubPullRequestUpsert  ToolName = "github_pull_request_upsert"
	ToolGitHubIssueCommentCreate ToolName = "github_issue_comment_create"
	ToolGitHubLabelsAdd          ToolName = "github_labels_add"
	ToolGitHubLabelsRemove       ToolName = "github_labels_remove"
)

const (
	ToolKubernetesPodsList       ToolName = "k8s_pods_list"
	ToolKubernetesEventsList     ToolName = "k8s_events_list"
	ToolKubernetesPodLogsGet     ToolName = "k8s_pod_logs_get"
	ToolKubernetesPodExec        ToolName = "k8s_pod_exec"
	ToolKubernetesManifestApply  ToolName = "k8s_manifest_apply"
	ToolKubernetesManifestDelete ToolName = "k8s_manifest_delete"
)

// ToolCategory marks read/write class used by policy and audit.
type ToolCategory string

const (
	ToolCategoryRead  ToolCategory = "read"
	ToolCategoryWrite ToolCategory = "write"
)

// ToolApprovalPolicy defines approval requirement for a tool.
type ToolApprovalPolicy string

const (
	ToolApprovalNone     ToolApprovalPolicy = "none"
	ToolApprovalRequired ToolApprovalPolicy = "required"
)

// ToolExecutionStatus is a normalized result status returned by tools.
type ToolExecutionStatus string

const (
	ToolExecutionStatusOK               ToolExecutionStatus = "ok"
	ToolExecutionStatusApprovalRequired ToolExecutionStatus = "approval_required"
)

// ToolCapability describes one tool in runtime catalog.
type ToolCapability struct {
	Name        ToolName           `json:"name"`
	Description string             `json:"description"`
	Category    ToolCategory       `json:"category"`
	Approval    ToolApprovalPolicy `json:"approval"`
}

// SessionContext is an authenticated MCP session bound to one run.
type SessionContext struct {
	RunID         string
	CorrelationID string
	ProjectID     string
	Namespace     string
	RuntimeMode   agentdomain.RuntimeMode
	ExpiresAt     time.Time
}

// IssueRunTokenParams describes token issuance request for one run.
type IssueRunTokenParams struct {
	RunID       string
	Namespace   string
	RuntimeMode agentdomain.RuntimeMode
	TTL         time.Duration
}

// IssuedToken holds issued bearer token metadata.
type IssuedToken struct {
	Token     string
	ExpiresAt time.Time
}

// PromptContext is deterministic render context for final prompt assembly.
type PromptContext struct {
	Version     string                   `json:"version"`
	Run         PromptRunContext         `json:"run"`
	Repository  PromptRepositoryContext  `json:"repository"`
	Issue       *PromptIssueContext      `json:"issue,omitempty"`
	Environment PromptEnvironmentContext `json:"environment"`
	Services    []PromptServiceContext   `json:"services"`
	MCP         PromptMCPContext         `json:"mcp"`
}

// PromptRunContext contains run/session identifiers for prompt render.
type PromptRunContext struct {
	RunID         string                  `json:"run_id"`
	CorrelationID string                  `json:"correlation_id"`
	ProjectID     string                  `json:"project_id"`
	Namespace     string                  `json:"namespace,omitempty"`
	RuntimeMode   agentdomain.RuntimeMode `json:"runtime_mode"`
}

// PromptRepositoryContext contains repository metadata for current run.
type PromptRepositoryContext struct {
	Provider     string `json:"provider"`
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	FullName     string `json:"full_name"`
	ServicesYAML string `json:"services_yaml"`
}

// PromptIssueContext contains issue metadata from run payload.
type PromptIssueContext struct {
	Number int64  `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url,omitempty"`
}

// PromptEnvironmentContext contains environment metadata.
type PromptEnvironmentContext struct {
	ServiceName string `json:"service_name"`
	MCPBaseURL  string `json:"mcp_base_url"`
}

// PromptServiceContext describes one platform service useful for prompt context.
type PromptServiceContext struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Kind     string `json:"kind"`
}

// PromptMCPContext describes tool catalog and policy flags.
type PromptMCPContext struct {
	ServerName string           `json:"server_name"`
	Tools      []ToolCapability `json:"tools"`
}

// ApprovalRequiredResult is returned by tools that require approval.
type ApprovalRequiredResult struct {
	Status  ToolExecutionStatus `json:"status"`
	Tool    ToolName            `json:"tool"`
	Message string              `json:"message"`
}

// PromptContextResult is output for prompt context tool/resource.
type PromptContextResult struct {
	Status  ToolExecutionStatus `json:"status"`
	Context PromptContext       `json:"context"`
}

// GitHubIssueGetInput describes issue lookup input.
type GitHubIssueGetInput struct {
	IssueNumber int `json:"issue_number,omitempty"`
}

// GitHubPullRequestGetInput describes pull request lookup input.
type GitHubPullRequestGetInput struct {
	PullRequestNumber int `json:"pull_request_number"`
}

// GitHubIssueCommentsListInput describes issue comments list input.
type GitHubIssueCommentsListInput struct {
	IssueNumber int `json:"issue_number,omitempty"`
	Limit       int `json:"limit,omitempty"`
}

// GitHubLabelsListInput describes issue labels list input.
type GitHubLabelsListInput struct {
	IssueNumber int `json:"issue_number,omitempty"`
}

// GitHubBranchesListInput describes branches list input.
type GitHubBranchesListInput struct {
	Limit int `json:"limit,omitempty"`
}

// GitHubBranchEnsureInput describes branch create/sync input.
type GitHubBranchEnsureInput struct {
	BranchName string `json:"branch_name"`
	BaseBranch string `json:"base_branch,omitempty"`
	BaseSHA    string `json:"base_sha,omitempty"`
	Force      bool   `json:"force,omitempty"`
}

// GitHubPullRequestUpsertInput describes create/update PR input.
type GitHubPullRequestUpsertInput struct {
	PullRequestNumber int    `json:"pull_request_number,omitempty"`
	Title             string `json:"title"`
	Body              string `json:"body,omitempty"`
	HeadBranch        string `json:"head_branch"`
	BaseBranch        string `json:"base_branch,omitempty"`
	Draft             bool   `json:"draft,omitempty"`
}

// GitHubIssueCommentCreateInput describes issue/PR comment create input.
type GitHubIssueCommentCreateInput struct {
	IssueNumber int    `json:"issue_number,omitempty"`
	Body        string `json:"body"`
}

// GitHubLabelsAddInput describes add-labels input.
type GitHubLabelsAddInput struct {
	IssueNumber int      `json:"issue_number,omitempty"`
	Labels      []string `json:"labels"`
}

// GitHubLabelsRemoveInput describes remove-labels input.
type GitHubLabelsRemoveInput struct {
	IssueNumber int      `json:"issue_number,omitempty"`
	Labels      []string `json:"labels"`
}

// KubernetesPodsListInput describes pod list input.
type KubernetesPodsListInput struct {
	Limit int `json:"limit,omitempty"`
}

// KubernetesEventsListInput describes event list input.
type KubernetesEventsListInput struct {
	Limit int `json:"limit,omitempty"`
}

// KubernetesPodLogsGetInput describes pod logs input.
type KubernetesPodLogsGetInput struct {
	Pod       string `json:"pod"`
	Container string `json:"container,omitempty"`
	TailLines int64  `json:"tail_lines,omitempty"`
}

// KubernetesPodExecInput describes pod exec input.
type KubernetesPodExecInput struct {
	Pod       string   `json:"pod"`
	Container string   `json:"container,omitempty"`
	Command   []string `json:"command"`
}

// KubernetesManifestApplyInput describes manifest apply request.
type KubernetesManifestApplyInput struct {
	ManifestYAML string `json:"manifest_yaml"`
}

// KubernetesManifestDeleteInput describes manifest delete request.
type KubernetesManifestDeleteInput struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// GitHubIssue describes normalized issue details.
type GitHubIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

// GitHubPullRequest describes normalized pull request details.
type GitHubPullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
	Head   string `json:"head"`
	Base   string `json:"base"`
}

// GitHubIssueComment describes normalized issue comment details.
type GitHubIssueComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
	URL  string `json:"url"`
	User string `json:"user"`
}

// GitHubLabel describes normalized label details.
type GitHubLabel struct {
	Name string `json:"name"`
}

// GitHubBranch describes normalized branch details.
type GitHubBranch struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
}

// KubernetesPod describes pod list item.
type KubernetesPod struct {
	Name      string `json:"name"`
	Phase     string `json:"phase"`
	NodeName  string `json:"node_name,omitempty"`
	StartTime string `json:"start_time,omitempty"`
}

// KubernetesEvent describes event list item.
type KubernetesEvent struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Object    string `json:"object"`
	Timestamp string `json:"timestamp"`
}

// KubernetesExecResult describes exec output.
type KubernetesExecResult struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr,omitempty"`
}

// GitHubIssueGetResult is output for issue lookup tool.
type GitHubIssueGetResult struct {
	Status ToolExecutionStatus `json:"status"`
	Issue  GitHubIssue         `json:"issue"`
}

// GitHubPullRequestGetResult is output for PR lookup tool.
type GitHubPullRequestGetResult struct {
	Status      ToolExecutionStatus `json:"status"`
	PullRequest GitHubPullRequest   `json:"pull_request"`
}

// GitHubIssueCommentsListResult is output for comments list tool.
type GitHubIssueCommentsListResult struct {
	Status   ToolExecutionStatus  `json:"status"`
	Comments []GitHubIssueComment `json:"comments"`
}

// GitHubLabelsListResult is output for labels list tool.
type GitHubLabelsListResult struct {
	Status ToolExecutionStatus `json:"status"`
	Labels []GitHubLabel       `json:"labels"`
}

// GitHubBranchesListResult is output for branches list tool.
type GitHubBranchesListResult struct {
	Status   ToolExecutionStatus `json:"status"`
	Branches []GitHubBranch      `json:"branches"`
}

// GitHubBranchEnsureResult is output for branch ensure tool.
type GitHubBranchEnsureResult struct {
	Status  ToolExecutionStatus `json:"status"`
	Branch  GitHubBranch        `json:"branch"`
	Message string              `json:"message,omitempty"`
}

// GitHubPullRequestUpsertResult is output for PR upsert tool.
type GitHubPullRequestUpsertResult struct {
	Status      ToolExecutionStatus `json:"status"`
	PullRequest GitHubPullRequest   `json:"pull_request"`
	Message     string              `json:"message,omitempty"`
}

// GitHubIssueCommentCreateResult is output for comment create tool.
type GitHubIssueCommentCreateResult struct {
	Status  ToolExecutionStatus `json:"status"`
	Comment GitHubIssueComment  `json:"comment"`
	Message string              `json:"message,omitempty"`
}

// GitHubLabelsMutationResult is output for labels add/remove tools.
type GitHubLabelsMutationResult struct {
	Status  ToolExecutionStatus `json:"status"`
	Labels  []GitHubLabel       `json:"labels"`
	Message string              `json:"message,omitempty"`
}

// KubernetesPodsListResult is output for pods list tool.
type KubernetesPodsListResult struct {
	Status ToolExecutionStatus `json:"status"`
	Pods   []KubernetesPod     `json:"pods"`
}

// KubernetesEventsListResult is output for events list tool.
type KubernetesEventsListResult struct {
	Status ToolExecutionStatus `json:"status"`
	Events []KubernetesEvent   `json:"events"`
}

// KubernetesPodLogsGetResult is output for pod logs tool.
type KubernetesPodLogsGetResult struct {
	Status ToolExecutionStatus `json:"status"`
	Logs   string              `json:"logs"`
}

// KubernetesPodExecToolResult is output for pod exec tool.
type KubernetesPodExecToolResult struct {
	Status  ToolExecutionStatus  `json:"status"`
	Exec    KubernetesExecResult `json:"exec"`
	Message string               `json:"message,omitempty"`
}
