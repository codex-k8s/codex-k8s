package worker

import (
	"context"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

// JobState is an external run workload state in Kubernetes.
type JobState string

const (
	// JobStatePending indicates workload exists but is not completed.
	JobStatePending JobState = "pending"
	// JobStateRunning indicates workload is actively running.
	JobStateRunning JobState = "running"
	// JobStateSucceeded indicates workload completed successfully.
	JobStateSucceeded JobState = "succeeded"
	// JobStateFailed indicates workload completed with failure.
	JobStateFailed JobState = "failed"
	// JobStateNotFound indicates workload was not found.
	JobStateNotFound JobState = "not_found"
)

// JobRef identifies a Kubernetes run workload.
type JobRef struct {
	// Namespace is a Kubernetes namespace where Job exists.
	Namespace string
	// Name is a Kubernetes Job resource name.
	Name string
}

// NamespaceSpec defines namespace baseline that should exist before launching a run job.
type NamespaceSpec struct {
	// RunID identifies run owning namespace lifecycle.
	RunID string
	// ProjectID identifies project scope for namespace metadata.
	ProjectID string
	// CorrelationID links namespace events to webhook flow.
	CorrelationID string
	// RuntimeMode controls whether namespace should be managed.
	RuntimeMode agentdomain.RuntimeMode
	// Namespace is target namespace name.
	Namespace string
}

// JobSpec defines parameters for Job launch.
type JobSpec struct {
	// RunID is a unique run identifier.
	RunID string
	// CorrelationID links workload to webhook flow.
	CorrelationID string
	// ProjectID is an effective project scope.
	ProjectID string
	// SlotNo is a leased slot number.
	SlotNo int
	// RuntimeMode defines execution profile for run.
	RuntimeMode agentdomain.RuntimeMode
	// Namespace is target namespace for Job.
	Namespace string
}

// Launcher creates and reconciles Kubernetes Jobs for runs.
type Launcher interface {
	// JobRef builds deterministic Job reference for run id.
	JobRef(runID string, namespace string) JobRef
	// EnsureNamespace prepares namespace baseline for full-env execution.
	EnsureNamespace(ctx context.Context, spec NamespaceSpec) error
	// CleanupNamespace removes runtime namespace after run completion.
	CleanupNamespace(ctx context.Context, spec NamespaceSpec) error
	// Launch creates Job if needed and returns its reference.
	Launch(ctx context.Context, spec JobSpec) (JobRef, error)
	// Status returns current workload state for a given Job reference.
	Status(ctx context.Context, ref JobRef) (JobState, error)
}
