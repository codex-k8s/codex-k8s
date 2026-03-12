package staff

import (
	"context"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	"github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	nextstepdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/nextstep"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/learningfeedback"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	projecttokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projecttoken"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	runtimeerrorrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimeerror"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// Config defines staff service behavior.
type Config struct {
	// LearningModeDefault is the default for newly created projects.
	LearningModeDefault bool

	// WebhookSpec is used when attaching repositories to projects.
	WebhookSpec provider.WebhookSpec

	// ProtectedProjectIDs is a set of project ids that must never be deleted via staff API.
	ProtectedProjectIDs map[string]struct{}
	// ProtectedRepositoryIDs is a set of repository binding ids that must never be deleted via staff API.
	ProtectedRepositoryIDs map[string]struct{}
	// PromptSeedsDir points to runner prompt seeds used for bootstrap sync.
	PromptSeedsDir string
	// NextStepLabels defines env-aware run:* labels allowed for next-step actions.
	NextStepLabels nextstepdomain.Labels
}

// Service exposes staff-only read/write operations protected by JWT + RBAC.
type Service struct {
	cfg           Config
	users         userrepo.Repository
	projects      projectrepo.Repository
	members       projectmemberrepo.Repository
	repos         repocfgrepo.Repository
	projectTokens projecttokenrepo.Repository
	feedback      learningfeedbackrepo.Repository
	runs          staffrunrepo.Repository
	tasks         runtimedeploytaskrepo.Repository
	runtimeErrors runtimeerrorrepo.Repository
	k8s           kubernetesConfigSync

	tokencrypt     *tokencrypt.Service
	platformTokens platformTokensRepository
	github         provider.RepositoryProvider
	githubMgmt     githubManagementClient
	runStatus      runNamespaceService
	runtimeDeploy  runtimeDeployController
}

type platformTokensRepository interface {
	Get(ctx context.Context) (entitytypes.PlatformGitHubTokens, bool, error)
}

type githubManagementClient interface {
	Preflight(ctx context.Context, params valuetypes.GitHubPreflightParams) (valuetypes.GitHubPreflightReport, error)
	GetDefaultBranch(ctx context.Context, token string, owner string, repo string) (string, error)
	GetFile(ctx context.Context, token string, owner string, repo string, path string, ref string) ([]byte, bool, error)
	CreatePullRequestWithFiles(ctx context.Context, token string, owner string, repo string, baseBranch string, headBranch string, title string, body string, files map[string][]byte) (prNumber int, prURL string, err error)
	ListIssueLabels(ctx context.Context, token string, owner string, repo string, issueNumber int) ([]string, error)
	AddIssueLabels(ctx context.Context, token string, owner string, repo string, issueNumber int, labels []string) ([]string, error)
	RemoveIssueLabel(ctx context.Context, token string, owner string, repo string, issueNumber int, label string) error
}

type kubernetesConfigSync interface {
	ListSecretNames(ctx context.Context, namespace string) ([]string, error)
	ListConfigMapNames(ctx context.Context, namespace string) ([]string, error)
	GetSecretData(ctx context.Context, namespace string, name string) (map[string][]byte, bool, error)
	UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error
	GetConfigMapData(ctx context.Context, namespace string, name string) (map[string]string, bool, error)
	UpsertConfigMap(ctx context.Context, namespace string, name string, data map[string]string) error
}

type runtimeDeployController interface {
	RequestTaskAction(ctx context.Context, params runtimedeploydomain.TaskActionParams) (runtimedeploydomain.TaskActionResult, error)
}

// Dependencies defines external collaborators required by staff service.
type Dependencies struct {
	Users          userrepo.Repository
	Projects       projectrepo.Repository
	Members        projectmemberrepo.Repository
	Repos          repocfgrepo.Repository
	ProjectTokens  projecttokenrepo.Repository
	Feedback       learningfeedbackrepo.Repository
	Runs           staffrunrepo.Repository
	Tasks          runtimedeploytaskrepo.Repository
	RuntimeErrors  runtimeerrorrepo.Repository
	K8s            kubernetesConfigSync
	Tokencrypt     *tokencrypt.Service
	PlatformTokens platformTokensRepository
	GitHub         provider.RepositoryProvider
	GitHubMgmt     githubManagementClient
	RunStatus      runNamespaceService
	RuntimeDeploy  runtimeDeployController
}

// NewService constructs staff service.
func NewService(cfg Config, deps Dependencies) *Service {
	return &Service{
		cfg:            cfg,
		users:          deps.Users,
		projects:       deps.Projects,
		members:        deps.Members,
		repos:          deps.Repos,
		projectTokens:  deps.ProjectTokens,
		feedback:       deps.Feedback,
		runs:           deps.Runs,
		tasks:          deps.Tasks,
		runtimeErrors:  deps.RuntimeErrors,
		k8s:            deps.K8s,
		tokencrypt:     deps.Tokencrypt,
		platformTokens: deps.PlatformTokens,
		github:         deps.GitHub,
		githubMgmt:     deps.GitHubMgmt,
		runStatus:      deps.RunStatus,
		runtimeDeploy:  deps.RuntimeDeploy,
	}
}
