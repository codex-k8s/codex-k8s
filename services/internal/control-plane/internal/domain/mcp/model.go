package mcp

import (
	"encoding/json"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

// ToolName is a stable MCP tool identifier.
type ToolName string

const ToolPromptContextGet ToolName = "codex_prompt_context_get"

const (
	ToolMCPSecretSyncEnv        ToolName = "secret.sync.github_k8s"
	ToolMCPDatabaseLifecycle    ToolName = "database.lifecycle"
	ToolMCPOwnerFeedbackRequest ToolName = "owner.feedback.request"
	ToolSelfImproveRunsList     ToolName = "self_improve_runs_list"
	ToolSelfImproveRunLookup    ToolName = "self_improve_run_lookup"
	ToolSelfImproveSessionGet   ToolName = "self_improve_session_get"
)

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
	ToolGitHubLabelsTransition   ToolName = "github_labels_transition"
)

const (
	ToolKubernetesPodsList                     ToolName = "k8s_pods_list"
	ToolKubernetesEventsList                   ToolName = "k8s_events_list"
	ToolKubernetesDeploymentsList              ToolName = "k8s_deployments_list"
	ToolKubernetesDaemonSetsList               ToolName = "k8s_daemonsets_list"
	ToolKubernetesStatefulSetsList             ToolName = "k8s_statefulsets_list"
	ToolKubernetesReplicaSetsList              ToolName = "k8s_replicasets_list"
	ToolKubernetesReplicationControllersList   ToolName = "k8s_replicationcontrollers_list"
	ToolKubernetesJobsList                     ToolName = "k8s_jobs_list"
	ToolKubernetesCronJobsList                 ToolName = "k8s_cronjobs_list"
	ToolKubernetesConfigMapsList               ToolName = "k8s_configmaps_list"
	ToolKubernetesSecretsList                  ToolName = "k8s_secrets_list"
	ToolKubernetesResourceQuotasList           ToolName = "k8s_resourcequotas_list"
	ToolKubernetesHorizontalPodAutoscalersList ToolName = "k8s_hpas_list"
	ToolKubernetesServicesList                 ToolName = "k8s_services_list"
	ToolKubernetesEndpointsList                ToolName = "k8s_endpoints_list"
	ToolKubernetesIngressesList                ToolName = "k8s_ingresses_list"
	ToolKubernetesIngressClassesList           ToolName = "k8s_ingressclasses_list"
	ToolKubernetesNetworkPoliciesList          ToolName = "k8s_networkpolicies_list"
	ToolKubernetesPersistentVolumeClaimsList   ToolName = "k8s_pvcs_list"
	ToolKubernetesPersistentVolumesList        ToolName = "k8s_pvs_list"
	ToolKubernetesStorageClassesList           ToolName = "k8s_storageclasses_list"
	ToolKubernetesPodLogsGet                   ToolName = "k8s_pod_logs_get"
	ToolKubernetesPodExec                      ToolName = "k8s_pod_exec"
	ToolKubernetesPodPortForward               ToolName = "k8s_pod_port_forward"
	ToolKubernetesManifestApply                ToolName = "k8s_manifest_apply"
	ToolKubernetesManifestDelete               ToolName = "k8s_manifest_delete"
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
	ToolApprovalNone      ToolApprovalPolicy = "none"
	ToolApprovalOwner     ToolApprovalPolicy = "owner"
	ToolApprovalDelegated ToolApprovalPolicy = "delegated"
	ToolApprovalRequired  ToolApprovalPolicy = ToolApprovalOwner
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
	Role        PromptRoleContext        `json:"role"`
	Docs        []PromptProjectDocRef    `json:"docs,omitempty"`
	Environment PromptEnvironmentContext `json:"environment"`
	Runtime     PromptRuntimeContext     `json:"runtime"`
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

// PromptRoleContext describes agent role-specific responsibilities and limits.
type PromptRoleContext struct {
	AgentKey     string                 `json:"agent_key,omitempty"`
	DisplayName  string                 `json:"display_name,omitempty"`
	Capabilities []PromptRoleCapability `json:"capabilities,omitempty"`
}

// PromptRoleCapability describes one role capability domain.
type PromptRoleCapability struct {
	Area  string `json:"area"`
	Scope string `json:"scope"`
	Notes string `json:"notes,omitempty"`
}

// PromptProjectDocRef is one role-filtered docs tree entry exported from services.yaml.
type PromptProjectDocRef struct {
	Path        string   `json:"path"`
	Description string   `json:"description,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Optional    bool     `json:"optional,omitempty"`
}

// PromptEnvironmentContext contains environment metadata.
type PromptEnvironmentContext struct {
	ServiceName string `json:"service_name"`
	MCPBaseURL  string `json:"mcp_base_url"`
}

// PromptRuntimeContext describes resolved runtime/deploy hints for current run.
type PromptRuntimeContext struct {
	TargetEnv       string                        `json:"target_env"`
	ServicesYAML    string                        `json:"services_yaml"`
	InventorySource string                        `json:"inventory_source,omitempty"`
	Inventory       []PromptRuntimeServiceContext `json:"inventory,omitempty"`
	Hints           PromptRuntimeHints            `json:"hints"`
}

// PromptRuntimeServiceContext describes one deploy service resolved from services.yaml.
type PromptRuntimeServiceContext struct {
	Name               string   `json:"name"`
	DeployGroup        string   `json:"deploy_group,omitempty"`
	CodeUpdateStrategy string   `json:"code_update_strategy"`
	DependsOn          []string `json:"depends_on,omitempty"`
	ManifestPaths      []string `json:"manifest_paths,omitempty"`
}

// PromptRuntimeHints describes runtime context metadata used by the agent.
type PromptRuntimeHints struct {
	RuntimeMode     agentdomain.RuntimeMode `json:"runtime_mode"`
	RepositoryRoot  string                  `json:"repository_root,omitempty"`
	InventoryLoaded bool                    `json:"inventory_loaded"`
	InventoryError  string                  `json:"inventory_error,omitempty"`
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
	IssueNumber               int  `json:"issue_number,omitempty"`
	Limit                     int  `json:"limit,omitempty"`
	IncludeTokenOwnerComments bool `json:"include_token_owner_comments,omitempty"`
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

// GitHubLabelsTransitionInput describes one labels transition request.
type GitHubLabelsTransitionInput struct {
	IssueNumber  int      `json:"issue_number,omitempty"`
	RemoveLabels []string `json:"remove_labels,omitempty"`
	AddLabels    []string `json:"add_labels,omitempty"`
}

// KubernetesPodsListInput describes pod list input.
type KubernetesPodsListInput struct {
	Limit int `json:"limit,omitempty"`
}

// KubernetesEventsListInput describes event list input.
type KubernetesEventsListInput struct {
	Limit int `json:"limit,omitempty"`
}

// KubernetesResourceListInput describes generic list input for namespace resources.
type KubernetesResourceListInput struct {
	Kind  KubernetesResourceKind `json:"kind,omitempty"`
	Limit int                    `json:"limit,omitempty"`
}

// KubernetesResourceKind identifies one supported Kubernetes resource class for list tools.
type KubernetesResourceKind string

const (
	KubernetesResourceKindDeployment            KubernetesResourceKind = "deployment"
	KubernetesResourceKindDaemonSet             KubernetesResourceKind = "daemonset"
	KubernetesResourceKindStatefulSet           KubernetesResourceKind = "statefulset"
	KubernetesResourceKindReplicaSet            KubernetesResourceKind = "replicaset"
	KubernetesResourceKindReplicationController KubernetesResourceKind = "replicationcontroller"
	KubernetesResourceKindJob                   KubernetesResourceKind = "job"
	KubernetesResourceKindCronJob               KubernetesResourceKind = "cronjob"
	KubernetesResourceKindConfigMap             KubernetesResourceKind = "configmap"
	KubernetesResourceKindSecret                KubernetesResourceKind = "secret"
	KubernetesResourceKindResourceQuota         KubernetesResourceKind = "resourcequota"
	KubernetesResourceKindHPA                   KubernetesResourceKind = "horizontalpodautoscaler"
	KubernetesResourceKindService               KubernetesResourceKind = "service"
	KubernetesResourceKindEndpoints             KubernetesResourceKind = "endpoints"
	KubernetesResourceKindIngress               KubernetesResourceKind = "ingress"
	KubernetesResourceKindIngressClass          KubernetesResourceKind = "ingressclass"
	KubernetesResourceKindNetworkPolicy         KubernetesResourceKind = "networkpolicy"
	KubernetesResourceKindPVC                   KubernetesResourceKind = "persistentvolumeclaim"
	KubernetesResourceKindPV                    KubernetesResourceKind = "persistentvolume"
	KubernetesResourceKindStorageClass          KubernetesResourceKind = "storageclass"
)

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

// KubernetesPodPortForwardInput describes pod port-forward request.
type KubernetesPodPortForwardInput struct {
	Pod        string `json:"pod"`
	Container  string `json:"container,omitempty"`
	LocalPort  int32  `json:"local_port"`
	RemotePort int32  `json:"remote_port"`
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

// GitHubIssueReaction describes normalized issue reaction details.
type GitHubIssueReaction struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
	User    string `json:"user"`
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

// KubernetesResourceRef describes one Kubernetes object in list-like tools.
type KubernetesResourceRef struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
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

// KubernetesResourceListResult is output for generic Kubernetes resource list tools.
type KubernetesResourceListResult struct {
	Status ToolExecutionStatus     `json:"status"`
	Items  []KubernetesResourceRef `json:"items"`
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

// KubernetesPodPortForwardResult is output for pod port-forward tool.
type KubernetesPodPortForwardResult struct {
	Status  ToolExecutionStatus `json:"status"`
	Message string              `json:"message,omitempty"`
}

// SecretSyncPolicy describes secret generation behavior for sync requests.
type SecretSyncPolicy string

const (
	SecretSyncPolicyDeterministic SecretSyncPolicy = "deterministic"
	SecretSyncPolicyRandom        SecretSyncPolicy = "random"
	SecretSyncPolicyProvided      SecretSyncPolicy = "provided"
)

// SecretSyncEnvInput describes deterministic secret sync request across GitHub and Kubernetes.
type SecretSyncEnvInput struct {
	ProjectID            string           `json:"project_id,omitempty"`
	Repository           string           `json:"repository,omitempty"`
	Environment          string           `json:"environment"`
	GitHubSecretName     string           `json:"github_secret_name"`
	KubernetesNamespace  string           `json:"kubernetes_namespace,omitempty"`
	KubernetesSecretName string           `json:"kubernetes_secret_name"`
	KubernetesSecretKey  string           `json:"kubernetes_secret_key,omitempty"`
	Policy               SecretSyncPolicy `json:"policy,omitempty"`
	SecretValue          string           `json:"secret_value,omitempty"`
	IdempotencyKey       string           `json:"idempotency_key,omitempty"`
	DryRun               bool             `json:"dry_run,omitempty"`
}

// SecretSyncEnvResult is output for secret.sync.github_k8s tool.
type SecretSyncEnvResult struct {
	Status         ToolExecutionStatus `json:"status"`
	RequestID      int64               `json:"request_id,omitempty"`
	ApprovalState  string              `json:"approval_state,omitempty"`
	Environment    string              `json:"environment,omitempty"`
	GitHubSecret   string              `json:"github_secret,omitempty"`
	KubernetesRef  string              `json:"kubernetes_ref,omitempty"`
	Policy         string              `json:"policy,omitempty"`
	IdempotencyKey string              `json:"idempotency_key,omitempty"`
	Reused         bool                `json:"reused,omitempty"`
	DryRun         bool                `json:"dry_run,omitempty"`
	Message        string              `json:"message,omitempty"`
}

// DatabaseLifecycleAction defines supported database lifecycle actions.
type DatabaseLifecycleAction string

const (
	DatabaseLifecycleActionCreate   DatabaseLifecycleAction = "create"
	DatabaseLifecycleActionDelete   DatabaseLifecycleAction = "delete"
	DatabaseLifecycleActionDescribe DatabaseLifecycleAction = "describe"
)

// DatabaseLifecycleInput describes database lifecycle request.
type DatabaseLifecycleInput struct {
	Environment   string                  `json:"environment"`
	Action        DatabaseLifecycleAction `json:"action"`
	DatabaseName  string                  `json:"database_name"`
	ConfirmDelete bool                    `json:"confirm_delete,omitempty"`
	DryRun        bool                    `json:"dry_run,omitempty"`
}

// DatabaseLifecycleResult is output for database.lifecycle tool.
type DatabaseLifecycleResult struct {
	Status         ToolExecutionStatus `json:"status"`
	RequestID      int64               `json:"request_id,omitempty"`
	ApprovalState  string              `json:"approval_state,omitempty"`
	Environment    string              `json:"environment,omitempty"`
	Action         string              `json:"action,omitempty"`
	DatabaseName   string              `json:"database_name,omitempty"`
	Applied        bool                `json:"applied,omitempty"`
	Exists         bool                `json:"exists"`
	OwnedByProject bool                `json:"owned_by_project"`
	OwnerProjectID string              `json:"owner_project_id,omitempty"`
	DryRun         bool                `json:"dry_run,omitempty"`
	Message        string              `json:"message,omitempty"`
}

// OwnerFeedbackRequestInput describes owner feedback request with fixed options and optional custom answer.
type OwnerFeedbackRequestInput struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	AllowCustom bool     `json:"allow_custom,omitempty"`
	DryRun      bool     `json:"dry_run,omitempty"`
}

// OwnerFeedbackRequestResult is output for owner.feedback.request tool.
type OwnerFeedbackRequestResult struct {
	Status        ToolExecutionStatus `json:"status"`
	RequestID     int64               `json:"request_id,omitempty"`
	ApprovalState string              `json:"approval_state,omitempty"`
	Question      string              `json:"question,omitempty"`
	Options       []string            `json:"options,omitempty"`
	DryRun        bool                `json:"dry_run,omitempty"`
	Message       string              `json:"message,omitempty"`
}

// SelfImproveRunsListInput describes paginated run history request.
type SelfImproveRunsListInput struct {
	RepositoryFullName string `json:"repository_full_name,omitempty"`
	Page               int    `json:"page,omitempty"`
	Limit              int    `json:"limit,omitempty"`
}

// SelfImproveRunLookupInput describes run search by issue/pr references.
type SelfImproveRunLookupInput struct {
	RepositoryFullName string `json:"repository_full_name,omitempty"`
	IssueNumber        int64  `json:"issue_number,omitempty"`
	PullRequestNumber  int64  `json:"pull_request_number,omitempty"`
	Limit              int    `json:"limit,omitempty"`
}

// SelfImproveSessionGetInput describes codex session retrieval input.
type SelfImproveSessionGetInput struct {
	RunID string `json:"run_id"`
}

// SelfImproveRunRef describes one run item in self-improve diagnostics.
type SelfImproveRunRef struct {
	RunID              string `json:"run_id"`
	CorrelationID      string `json:"correlation_id"`
	ProjectID          string `json:"project_id,omitempty"`
	RepositoryFullName string `json:"repository_full_name,omitempty"`
	AgentKey           string `json:"agent_key,omitempty"`
	IssueNumber        int64  `json:"issue_number,omitempty"`
	IssueURL           string `json:"issue_url,omitempty"`
	PullRequestNumber  int64  `json:"pull_request_number,omitempty"`
	PullRequestURL     string `json:"pull_request_url,omitempty"`
	TriggerKind        string `json:"trigger_kind,omitempty"`
	TriggerLabel       string `json:"trigger_label,omitempty"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at,omitempty"`
	StartedAt          string `json:"started_at,omitempty"`
	FinishedAt         string `json:"finished_at,omitempty"`
}

// SelfImproveRunsListResult is output for paginated run history.
type SelfImproveRunsListResult struct {
	Status  ToolExecutionStatus `json:"status"`
	Page    int                 `json:"page"`
	Limit   int                 `json:"limit"`
	HasNext bool                `json:"has_next"`
	Items   []SelfImproveRunRef `json:"items"`
}

// SelfImproveRunLookupResult is output for run search by issue/pr references.
type SelfImproveRunLookupResult struct {
	Status ToolExecutionStatus `json:"status"`
	Items  []SelfImproveRunRef `json:"items"`
}

// SelfImproveSessionGetResult is output for codex session extraction.
type SelfImproveSessionGetResult struct {
	Status           ToolExecutionStatus `json:"status"`
	Run              SelfImproveRunRef   `json:"run"`
	TmpDirectory     string              `json:"tmp_directory"`
	TmpFilePath      string              `json:"tmp_file_path"`
	CodexSessionJSON json.RawMessage     `json:"codex_session_json"`
}

// ApprovalDecision describes external decision for one mcp_action_request.
type ApprovalDecision string

const (
	ApprovalDecisionApproved ApprovalDecision = "approved"
	ApprovalDecisionDenied   ApprovalDecision = "denied"
	ApprovalDecisionExpired  ApprovalDecision = "expired"
	ApprovalDecisionFailed   ApprovalDecision = "failed"
	ApprovalDecisionApplied  ApprovalDecision = "applied"
)

// ResolveApprovalParams describes one approval decision update.
type ResolveApprovalParams struct {
	RequestID int64
	Decision  ApprovalDecision
	ActorID   string
	Reason    string
}

// ApprovalListItem is staff-facing pending approval queue entry.
type ApprovalListItem struct {
	ID            int64     `json:"id"`
	CorrelationID string    `json:"correlation_id"`
	RunID         string    `json:"run_id,omitempty"`
	ProjectID     string    `json:"project_id,omitempty"`
	ProjectSlug   string    `json:"project_slug,omitempty"`
	ProjectName   string    `json:"project_name,omitempty"`
	IssueNumber   int       `json:"issue_number,omitempty"`
	PRNumber      int       `json:"pr_number,omitempty"`
	TriggerLabel  string    `json:"trigger_label,omitempty"`
	ToolName      string    `json:"tool_name"`
	Action        string    `json:"action"`
	ApprovalMode  string    `json:"approval_mode"`
	RequestedBy   string    `json:"requested_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// ResolveApprovalResult returns updated approval request summary.
type ResolveApprovalResult struct {
	ID            int64  `json:"id"`
	CorrelationID string `json:"correlation_id"`
	RunID         string `json:"run_id,omitempty"`
	ToolName      string `json:"tool_name"`
	Action        string `json:"action"`
	ApprovalState string `json:"approval_state"`
}
