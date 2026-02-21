package staff

import (
	"context"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	"github.com/codex-k8s/codex-k8s/libs/go/repo/provider"
	configentryrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/configentry"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/learningfeedback"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	projecttokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projecttoken"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	runtimeerrorrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimeerror"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
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
}

// Service exposes staff-only read/write operations protected by JWT + RBAC.
type Service struct {
	cfg           Config
	users         userrepo.Repository
	projects      projectrepo.Repository
	members       projectmemberrepo.Repository
	repos         repocfgrepo.Repository
	projectTokens projecttokenrepo.Repository
	configEntries configentryrepo.Repository
	feedback      learningfeedbackrepo.Repository
	runs          staffrunrepo.Repository
	tasks         runtimedeploytaskrepo.Repository
	runtimeErrors runtimeerrorrepo.Repository
	images        registryImageService
	k8s           kubernetesConfigSync

	tokencrypt     *tokencrypt.Service
	platformTokens platformTokensRepository
	github         provider.RepositoryProvider
	githubMgmt     githubManagementClient
	runStatus      runNamespaceService
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
	EnsureEnvironment(ctx context.Context, token string, owner string, repo string, envName string) error
	ListEnvSecretNames(ctx context.Context, token string, owner string, repo string, envName string) (map[string]struct{}, error)
	ListEnvVariableValues(ctx context.Context, token string, owner string, repo string, envName string) (map[string]string, error)
	UpsertEnvSecret(ctx context.Context, token string, owner string, repo string, envName string, key string, value string) error
	UpsertEnvVariable(ctx context.Context, token string, owner string, repo string, envName string, key string, value string) error
}

type registryImageService interface {
	List(ctx context.Context, filter querytypes.RegistryImageListFilter) ([]entitytypes.RegistryImageRepository, error)
	DeleteTag(ctx context.Context, params querytypes.RegistryImageDeleteParams) (entitytypes.RegistryImageDeleteResult, error)
	Cleanup(ctx context.Context, filter querytypes.RegistryImageCleanupFilter) (entitytypes.RegistryImageCleanupResult, error)
}

type kubernetesConfigSync interface {
	ListSecretNames(ctx context.Context, namespace string) ([]string, error)
	ListConfigMapNames(ctx context.Context, namespace string) ([]string, error)
	GetSecretData(ctx context.Context, namespace string, name string) (map[string][]byte, bool, error)
	UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error
	GetConfigMapData(ctx context.Context, namespace string, name string) (map[string]string, bool, error)
	UpsertConfigMap(ctx context.Context, namespace string, name string, data map[string]string) error
}

// NewService constructs staff service.
func NewService(
	cfg Config,
	users userrepo.Repository,
	projects projectrepo.Repository,
	members projectmemberrepo.Repository,
	repos repocfgrepo.Repository,
	projectTokens projecttokenrepo.Repository,
	configEntries configentryrepo.Repository,
	feedback learningfeedbackrepo.Repository,
	runs staffrunrepo.Repository,
	tasks runtimedeploytaskrepo.Repository,
	runtimeErrors runtimeerrorrepo.Repository,
	images registryImageService,
	k8s kubernetesConfigSync,
	tokencrypt *tokencrypt.Service,
	platformTokens platformTokensRepository,
	github provider.RepositoryProvider,
	githubMgmt githubManagementClient,
	runStatus runNamespaceService,
) *Service {
	return &Service{
		cfg:            cfg,
		users:          users,
		projects:       projects,
		members:        members,
		repos:          repos,
		projectTokens:  projectTokens,
		configEntries:  configEntries,
		feedback:       feedback,
		runs:           runs,
		tasks:          tasks,
		runtimeErrors:  runtimeErrors,
		images:         images,
		k8s:            k8s,
		tokencrypt:     tokencrypt,
		platformTokens: platformTokens,
		github:         github,
		githubMgmt:     githubMgmt,
		runStatus:      runStatus,
	}
}
