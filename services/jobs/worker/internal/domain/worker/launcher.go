package worker

import (
	"context"

	libslauncher "github.com/codex-k8s/codex-k8s/libs/go/k8s/joblauncher"
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
type JobSpec = libslauncher.JobSpec

// Launcher creates and reconciles Kubernetes Jobs for runs.
type Launcher interface {
	// JobRef builds deterministic Job reference for run id.
	JobRef(runID string, namespace string) JobRef
	// FindRunJobRefByRunID resolves Kubernetes Job reference by run-id label across namespaces.
	// Used when run job is created outside of the default full-env namespace strategy
	// (for example, inside a persistent slot namespace).
	FindRunJobRefByRunID(ctx context.Context, runID string) (JobRef, bool, error)
	// EnsureNamespace prepares namespace baseline for full-env execution.
	EnsureNamespace(ctx context.Context, spec NamespaceSpec) error
	// CleanupNamespace removes runtime namespace after run completion.
	CleanupNamespace(ctx context.Context, spec NamespaceSpec) error
	// Launch creates Job if needed and returns its reference.
	Launch(ctx context.Context, spec JobSpec) (JobRef, error)
	// Status returns current workload state for a given Job reference.
	Status(ctx context.Context, ref JobRef) (JobState, error)
}
