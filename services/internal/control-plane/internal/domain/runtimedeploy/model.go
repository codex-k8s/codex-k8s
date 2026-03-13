package runtimedeploy

import (
	"context"
	"log/slog"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/registry"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
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

// EvaluateReuseParams describes one reusable-namespace fast-path decision request.
type EvaluateReuseParams struct {
	RunID              string
	ProjectID          string
	IssueNumber        int64
	AgentKey           string
	RuntimeMode        string
	Namespace          string
	TargetEnv          string
	SlotNo             int
	RepositoryFullName string
	ServicesYAMLPath   string
	BuildRef           string
	DeployOnly         bool
}

// EvaluateReuseResult describes whether one managed namespace can skip runtime deploy/build.
type EvaluateReuseResult struct {
	Reusable          bool
	Namespace         string
	TargetEnv         string
	EffectiveBuildRef string
	FingerprintHash   string
	Reason            string
}

// RuntimeNamespaceState keeps managed namespace metadata required for fast-path reuse checks.
type RuntimeNamespaceState struct {
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	Terminating bool
}

// TaskAction identifies one operator-triggered runtime deploy control action.
type TaskAction = querytypes.RuntimeDeployTaskAction

const (
	TaskActionCancel TaskAction = querytypes.RuntimeDeployTaskActionCancel
	TaskActionStop   TaskAction = querytypes.RuntimeDeployTaskActionStop
)

// TaskActionActor identifies the staff principal who requested cancel/stop.
type TaskActionActor struct {
	UserID      string
	Email       string
	GitHubLogin string
}

// TaskActionParams describes one staff cancel/stop request.
type TaskActionParams struct {
	RunID       string
	Action      TaskAction
	Reason      string
	RequestedAt time.Time
	Actor       TaskActionActor
}

// TaskActionResult describes runtime deploy action outcome.
type TaskActionResult struct {
	RunID           string
	Action          TaskAction
	PreviousStatus  entitytypes.RuntimeDeployTaskStatus
	CurrentStatus   entitytypes.RuntimeDeployTaskStatus
	AlreadyTerminal bool
}

type flowEventRecorder interface {
	Insert(ctx context.Context, params floweventrepo.InsertParams) error
}

type runReader interface {
	GetByID(ctx context.Context, runID string) (agentrunrepo.Run, bool, error)
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
	GetManagedRunNamespace(ctx context.Context, namespace string) (RuntimeNamespaceState, bool, error)
	UpsertNamespaceAnnotations(ctx context.Context, namespace string, annotations map[string]string) error
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
	GetTagInfo(ctx context.Context, repository string, tag string) (registry.TagInfo, bool, error)
	ListTagInfos(ctx context.Context, repository string) ([]registry.TagInfo, error)
	DeleteTag(ctx context.Context, repository string, tag string) (registry.DeleteResult, error)
}

type runtimeErrorRecorder interface {
	RecordBestEffort(ctx context.Context, params querytypes.RuntimeErrorRecordParams)
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
	Runs       runReader
	FlowEvents flowEventRecorder
	Registry   RegistryClient
	RuntimeErr runtimeErrorRecorder
	Logger     *slog.Logger
}
