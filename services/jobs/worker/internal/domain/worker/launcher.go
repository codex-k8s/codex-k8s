package worker

import "context"

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
}

// Launcher creates and reconciles Kubernetes Jobs for runs.
type Launcher interface {
	// JobRef builds deterministic Job reference for run id.
	JobRef(runID string) JobRef
	// Launch creates Job if needed and returns its reference.
	Launch(ctx context.Context, spec JobSpec) (JobRef, error)
	// Status returns current workload state for a given Job reference.
	Status(ctx context.Context, ref JobRef) (JobState, error)
}
