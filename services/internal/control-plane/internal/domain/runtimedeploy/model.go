package runtimedeploy

import (
	"context"
	"log/slog"
	"time"
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
	ServicesConfigPath string
	RepositoryRoot     string
	RolloutTimeout     time.Duration
	KanikoTimeout      time.Duration
	KanikoFieldManager string
	GitHubPAT          string
}

// KubernetesClient describes Kubernetes operations used by runtime deploy orchestration.
type KubernetesClient interface {
	UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error
	UpsertConfigMap(ctx context.Context, namespace string, name string, data map[string]string) error
	GetSecretData(ctx context.Context, namespace string, name string) (map[string][]byte, bool, error)
	DeleteJobIfExists(ctx context.Context, namespace string, name string) error
	WaitForJobComplete(ctx context.Context, namespace string, name string, timeout time.Duration) error
	WaitForDeploymentReady(ctx context.Context, namespace string, name string, timeout time.Duration) error
	WaitForStatefulSetReady(ctx context.Context, namespace string, name string, timeout time.Duration) error
	WaitForDaemonSetReady(ctx context.Context, namespace string, name string, timeout time.Duration) error
	ApplyManifest(ctx context.Context, manifest []byte, namespaceOverride string, fieldManager string) ([]AppliedResourceRef, error)
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
	Logger     *slog.Logger
}
