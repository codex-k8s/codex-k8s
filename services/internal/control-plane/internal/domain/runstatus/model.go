package runstatus

import (
	"context"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	platformtokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/platformtoken"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

// Phase identifies a run status event reflected in one GitHub issue comment.
type Phase string

const (
	PhaseCreated          Phase = "created"
	PhaseStarted          Phase = "started"
	PhaseAuthRequired     Phase = "auth_required"
	PhaseAuthResolved     Phase = "auth_resolved"
	PhaseFinished         Phase = "finished"
	PhaseNamespaceDeleted Phase = "namespace_deleted"
)

// UpsertCommentParams describes one run status comment update request.
type UpsertCommentParams struct {
	RunID                    string
	Phase                    Phase
	JobName                  string
	JobNamespace             string
	RuntimeMode              string
	Namespace                string
	TriggerKind              string
	PromptLocale             string
	Model                    string
	ReasoningEffort          string
	RunStatus                string
	CodexAuthVerificationURL string
	CodexAuthUserCode        string
	Deleted                  bool
	AlreadyDeleted           bool
}

// UpsertCommentResult returns tracked issue comment metadata.
type UpsertCommentResult struct {
	CommentID  int64
	CommentURL string
}

// TriggerLabelConflictCommentParams describes one localized conflict comment request.
type TriggerLabelConflictCommentParams struct {
	CorrelationID      string
	RepositoryFullName string
	IssueNumber        int
	Locale             string
	TriggerLabel       string
	ConflictingLabels  []string
}

// TriggerLabelConflictCommentResult returns posted GitHub comment metadata.
type TriggerLabelConflictCommentResult struct {
	CommentID  int64
	CommentURL string
}

// TriggerWarningCommentParams describes one localized warning comment request
// when a webhook event was processed but run was not created.
type TriggerWarningCommentParams struct {
	CorrelationID      string
	RepositoryFullName string
	ThreadKind         string
	ThreadNumber       int
	Locale             string
	ReasonCode         TriggerWarningReasonCode
	ConflictingLabels  []string
}

// TriggerWarningCommentResult returns posted GitHub comment metadata.
type TriggerWarningCommentResult struct {
	CommentID  int64
	CommentURL string
}

// RequestedByType identifies who requested run namespace deletion.
type RequestedByType string

const (
	RequestedByTypeSystem    RequestedByType = "system"
	RequestedByTypeStaffUser RequestedByType = "staff_user"
)

// DeleteNamespaceParams describes one namespace delete request for a run.
type DeleteNamespaceParams struct {
	RunID           string
	RequestedByType RequestedByType
	RequestedByID   string
}

// DeleteNamespaceResult describes namespace delete operation outcome.
type DeleteNamespaceResult struct {
	RunID          string
	Namespace      string
	Deleted        bool
	AlreadyDeleted bool
	CommentURL     string
}

// RuntimeState describes current run runtime artifacts from status comment and Kubernetes.
type RuntimeState struct {
	HasStatusComment bool
	JobName          string
	JobNamespace     string
	Namespace        string
	JobExists        bool
	NamespaceExists  bool
}

// CleanupByIssueParams describes auto-cleanup scope for issue/pr close events.
type CleanupByIssueParams struct {
	RepositoryFullName string
	IssueNumber        int64
	RequestedByID      string
}

// CleanupByPullRequestParams describes auto-cleanup scope for pull request close/merge events.
type CleanupByPullRequestParams struct {
	RepositoryFullName string
	PRNumber           int64
	RequestedByID      string
}

// CleanupByIssueResult summarizes auto-cleanup outcomes.
type CleanupByIssueResult struct {
	MatchedRuns         int
	CleanedNamespaces   int
	AlreadyDeletedCount int
	SkippedRuns         int
	FailedRuns          int
}

// Config controls run status operations.
type Config struct {
	PublicBaseURL    string
	DefaultLocale    string
	AIDomain         string
	ProductionDomain string
}

// KubernetesClient provides namespace cleanup operation for runstatus service.
type KubernetesClient interface {
	DeleteManagedRunNamespace(ctx context.Context, namespace string) (bool, error)
	NamespaceExists(ctx context.Context, namespace string) (bool, error)
	JobExists(ctx context.Context, namespace string, jobName string) (bool, error)
	FindManagedRunNamespaceByRunID(ctx context.Context, runID string) (string, bool, error)
}

// GitHubClient provides issue comment operations for runstatus service.
type GitHubClient interface {
	ListIssueComments(ctx context.Context, params mcpdomain.GitHubListIssueCommentsParams) ([]mcpdomain.GitHubIssueComment, error)
	CreateIssueComment(ctx context.Context, params mcpdomain.GitHubCreateIssueCommentParams) (mcpdomain.GitHubIssueComment, error)
	EditIssueComment(ctx context.Context, params mcpdomain.GitHubEditIssueCommentParams) (mcpdomain.GitHubIssueComment, error)
}

// Dependencies wires required adapters for runstatus service.
type Dependencies struct {
	Runs       agentrunrepo.Repository
	Platform   platformtokenrepo.Repository
	TokenCrypt *tokencrypt.Service
	GitHub     GitHubClient
	Kubernetes KubernetesClient
	FlowEvents floweventrepo.Repository
}

// Service maintains one run status message in issue comments and handles forced namespace cleanup.
type Service struct {
	cfg Config

	runs       agentrunrepo.Repository
	platform   platformtokenrepo.Repository
	tokenCrypt *tokencrypt.Service
	github     GitHubClient
	kubernetes KubernetesClient
	flowEvents floweventrepo.Repository
}

type runContext struct {
	run                 agentrunrepo.Run
	payload             querytypes.RunPayload
	commentTargetNumber int
	commentTargetKind   commentTargetKind
	repoOwner           string
	repoName            string
	githubToken         string
	triggerKind         string
}

func (c runContext) hasCommentTarget() bool {
	return c.commentTargetNumber > 0 && strings.TrimSpace(string(c.commentTargetKind)) != ""
}

type commentState struct {
	RunID                    string `json:"run_id"`
	Phase                    Phase  `json:"phase"`
	JobName                  string `json:"job_name,omitempty"`
	JobNamespace             string `json:"job_namespace,omitempty"`
	RuntimeMode              string `json:"runtime_mode,omitempty"`
	Namespace                string `json:"namespace,omitempty"`
	SlotURL                  string `json:"slot_url,omitempty"`
	TriggerKind              string `json:"trigger_kind,omitempty"`
	PromptLocale             string `json:"prompt_locale,omitempty"`
	Model                    string `json:"model,omitempty"`
	ReasoningEffort          string `json:"reasoning_effort,omitempty"`
	RunStatus                string `json:"run_status,omitempty"`
	CodexAuthVerificationURL string `json:"codex_auth_verification_url,omitempty"`
	CodexAuthUserCode        string `json:"codex_auth_user_code,omitempty"`
	Deleted                  bool   `json:"deleted,omitempty"`
	AlreadyDeleted           bool   `json:"already_deleted,omitempty"`
}
