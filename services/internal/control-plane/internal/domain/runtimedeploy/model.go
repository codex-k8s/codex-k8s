package runtimedeploy

import (
	"context"
	"log/slog"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/registry"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
)

// PrepareParams describes one run-bound environment preparation request.
type PrepareParams struct {
	RunID              string
	RuntimeMode        string
	Namespace          string
	TargetEnv          string
	SlotNo             int
	RepositoryFullName string
	ServicesYAMLPath   string
	BuildRef           string
	DeployOnly         bool
}

// PrepareResult describes resolved runtime target after deployment preparation.
type PrepareResult struct {
	Namespace string
	TargetEnv string
}

// Config defines runtime deployment service options.
type Config struct {
	ServicesConfigPath      string
	RepositoryRoot          string
	RolloutTimeout          time.Duration
	KanikoTimeout           time.Duration
	WaitPollInterval        time.Duration
	KanikoFieldManager      string
	GitHubPAT               string
	RegistryCleanupKeepTags int
	KanikoJobLogTailLines   int64
}

// KubernetesClient describes Kubernetes operations used by runtime deploy orchestration.
type KubernetesClient interface {
	EnsureNamespace(ctx context.Context, namespace string) error
	UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error
	UpsertTLSSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error
	UpsertConfigMap(ctx context.Context, namespace string, name string, data map[string]string) error
	GetSecretData(ctx context.Context, namespace string, name string) (map[string][]byte, bool, error)
	DeleteJobIfExists(ctx context.Context, namespace string, name string) error
	WaitForJobComplete(ctx context.Context, namespace string, name string, timeout time.Duration) error
	GetJobLogs(ctx context.Context, namespace string, name string, tailLines int64) (string, error)
	WaitForDeploymentReady(ctx context.Context, namespace string, name string, timeout time.Duration) error
	WaitForStatefulSetReady(ctx context.Context, namespace string, name string, timeout time.Duration) error
	WaitForDaemonSetReady(ctx context.Context, namespace string, name string, timeout time.Duration) error
	ApplyManifest(ctx context.Context, manifest []byte, namespaceOverride string, fieldManager string) ([]AppliedResourceRef, error)
}

// RegistryClient describes internal registry operations required by runtime deploy.
type RegistryClient interface {
	ListTags(ctx context.Context, repository string) ([]string, error)
	ListTagInfos(ctx context.Context, repository string) ([]registry.TagInfo, error)
	DeleteTag(ctx context.Context, repository string, tag string) (registry.DeleteResult, error)
}

// AppliedResourceRef identifies one Kubernetes object applied from rendered manifest.
type AppliedResourceRef struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

// Dependencies wires runtime deployment collaborators.
type Dependencies struct {
	Kubernetes KubernetesClient
	Tasks      runtimedeploytaskrepo.Repository
	Registry   RegistryClient
	Logger     *slog.Logger
}
