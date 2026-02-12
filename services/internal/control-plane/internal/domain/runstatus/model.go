package runstatus

import (
	"context"

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
	PhaseStarted          Phase = "started"
	PhaseFinished         Phase = "finished"
	PhaseNamespaceDeleted Phase = "namespace_deleted"
)

// UpsertCommentParams describes one run status comment update request.
type UpsertCommentParams struct {
	RunID          string
	Phase          Phase
	JobName        string
	JobNamespace   string
	RuntimeMode    string
	Namespace      string
	TriggerKind    string
	PromptLocale   string
	RunStatus      string
	Deleted        bool
	AlreadyDeleted bool
}

// UpsertCommentResult returns tracked issue comment metadata.
type UpsertCommentResult struct {
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

// Config controls run status operations.
type Config struct {
	PublicBaseURL string
	DefaultLocale string
}

// KubernetesClient provides namespace cleanup operation for runstatus service.
type KubernetesClient interface {
	DeleteManagedRunNamespace(ctx context.Context, namespace string) (bool, error)
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
	run         agentrunrepo.Run
	payload     querytypes.RunPayload
	issueNumber int
	repoOwner   string
	repoName    string
	githubToken string
	triggerKind string
}

type commentState struct {
	RunID          string `json:"run_id"`
	Phase          Phase  `json:"phase"`
	JobName        string `json:"job_name,omitempty"`
	JobNamespace   string `json:"job_namespace,omitempty"`
	RuntimeMode    string `json:"runtime_mode,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	TriggerKind    string `json:"trigger_kind,omitempty"`
	PromptLocale   string `json:"prompt_locale,omitempty"`
	RunStatus      string `json:"run_status,omitempty"`
	Deleted        bool   `json:"deleted,omitempty"`
	AlreadyDeleted bool   `json:"already_deleted,omitempty"`
}
