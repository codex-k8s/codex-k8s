package worker

import (
	"context"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	libslauncher "github.com/codex-k8s/kodex/libs/go/k8s/joblauncher"
)

type JobState = libslauncher.JobState

const (
	JobStatePending   JobState = libslauncher.JobStatePending
	JobStateRunning   JobState = libslauncher.JobStateRunning
	JobStateSucceeded JobState = libslauncher.JobStateSucceeded
	JobStateFailed    JobState = libslauncher.JobStateFailed
	JobStateNotFound  JobState = libslauncher.JobStateNotFound
)

type JobRef = libslauncher.JobRef
type NamespaceSpec = libslauncher.NamespaceSpec
type NamespaceEnsureResult = libslauncher.NamespaceEnsureResult
type NamespaceReuseLookup = libslauncher.NamespaceReuseLookup
type NamespaceReuseResult = libslauncher.NamespaceReuseResult
type ManagedNamespaceState = libslauncher.ManagedNamespaceState
type ManagedNamespaceListParams = libslauncher.ManagedNamespaceListParams
type NamespaceWorkloadState = libslauncher.NamespaceWorkloadState
type JobSpec = libslauncher.JobSpec

// Launcher creates and reconciles Kubernetes run workloads (Job/Pod) for runs.
type Launcher interface {
	// JobRef builds deterministic workload reference for run id.
	JobRef(runID string, namespace string) JobRef
	// ListWorkerPodNames returns worker pod names currently visible in the given namespace.
	ListWorkerPodNames(ctx context.Context, namespace string) ([]string, error)
	// FindRunJobRefByRunID resolves Kubernetes Job reference by run-id label across namespaces.
	// Used when run job is created outside of the default full-env namespace strategy
	// (for example, inside a persistent slot namespace).
	FindRunJobRefByRunID(ctx context.Context, runID string) (JobRef, bool, error)
	// FindReusableNamespace resolves active namespace lease for one project/issue/agent tuple.
	FindReusableNamespace(ctx context.Context, lookup NamespaceReuseLookup) (NamespaceReuseResult, bool, error)
	// EnsureNamespace prepares namespace baseline for full-env execution.
	EnsureNamespace(ctx context.Context, spec NamespaceSpec) (NamespaceEnsureResult, error)
	// EnsureAccessProfile prepares ServiceAccount/RBAC profile in an existing namespace.
	EnsureAccessProfile(ctx context.Context, namespace string, profile agentdomain.RuntimeAccessProfile) (string, error)
	// ListManagedRunNamespaces returns worker-managed runtime namespaces for cleanup guardrails.
	ListManagedRunNamespaces(ctx context.Context, params ManagedNamespaceListParams) ([]ManagedNamespaceState, error)
	// InspectNamespaceWorkloads reports active workload objects inside one managed namespace.
	InspectNamespaceWorkloads(ctx context.Context, namespace string) (NamespaceWorkloadState, error)
	// DeleteManagedNamespace removes one worker-managed namespace after guardrails passed.
	DeleteManagedNamespace(ctx context.Context, namespace string) (bool, error)
	// Launch creates workload if needed and returns its reference.
	Launch(ctx context.Context, spec JobSpec) (JobRef, error)
	// Status returns current workload state for a given run workload reference.
	Status(ctx context.Context, ref JobRef) (JobState, error)
}
