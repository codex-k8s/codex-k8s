package launcher

import (
	"context"
	"fmt"

	libslauncher "github.com/codex-k8s/codex-k8s/libs/go/k8s/joblauncher"
	"github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/worker"
)

// Adapter bridges domain launcher port with client-go launcher implementation.
type Adapter struct {
	impl *libslauncher.Launcher
}

// NewAdapter creates domain-compatible Kubernetes launcher adapter.
func NewAdapter(cfg libslauncher.Config) (*Adapter, error) {
	impl, err := libslauncher.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes launcher: %w", err)
	}
	return &Adapter{impl: impl}, nil
}

// JobRef builds deterministic job reference for run id.
func (a *Adapter) JobRef(runID string) worker.JobRef {
	ref := a.impl.JobRef(runID)
	return worker.JobRef{Namespace: ref.Namespace, Name: ref.Name}
}

// Launch creates Kubernetes Job for run.
func (a *Adapter) Launch(ctx context.Context, spec worker.JobSpec) (worker.JobRef, error) {
	ref, err := a.impl.Launch(ctx, libslauncher.JobSpec{
		RunID:         spec.RunID,
		CorrelationID: spec.CorrelationID,
		ProjectID:     spec.ProjectID,
		SlotNo:        spec.SlotNo,
	})
	if err != nil {
		return worker.JobRef{}, err
	}
	return worker.JobRef{Namespace: ref.Namespace, Name: ref.Name}, nil
}

// Status returns current Kubernetes Job state.
func (a *Adapter) Status(ctx context.Context, ref worker.JobRef) (worker.JobState, error) {
	state, err := a.impl.Status(ctx, libslauncher.JobRef{Namespace: ref.Namespace, Name: ref.Name})
	if err != nil {
		return "", err
	}

	switch state {
	case libslauncher.JobStatePending:
		return worker.JobStatePending, nil
	case libslauncher.JobStateRunning:
		return worker.JobStateRunning, nil
	case libslauncher.JobStateSucceeded:
		return worker.JobStateSucceeded, nil
	case libslauncher.JobStateFailed:
		return worker.JobStateFailed, nil
	case libslauncher.JobStateNotFound:
		return worker.JobStateNotFound, nil
	default:
		return worker.JobStatePending, nil
	}
}
