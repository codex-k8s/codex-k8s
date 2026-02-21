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
func (a *Adapter) JobRef(runID string, namespace string) worker.JobRef {
	return a.impl.JobRef(runID, namespace)
}

func (a *Adapter) FindRunJobRefByRunID(ctx context.Context, runID string) (worker.JobRef, bool, error) {
	ref, ok, err := a.impl.FindRunJobRefByRunID(ctx, runID)
	if err != nil {
		return worker.JobRef{}, false, err
	}
	return ref, ok, nil
}

func (a *Adapter) FindReusableNamespace(ctx context.Context, lookup worker.NamespaceReuseLookup) (worker.NamespaceReuseResult, bool, error) {
	result, ok, err := a.impl.FindReusableNamespace(ctx, lookup)
	if err != nil {
		return worker.NamespaceReuseResult{}, false, err
	}
	return result, ok, nil
}

// EnsureNamespace prepares namespace baseline for full-env run.
func (a *Adapter) EnsureNamespace(ctx context.Context, spec worker.NamespaceSpec) (worker.NamespaceEnsureResult, error) {
	return a.impl.EnsureNamespace(ctx, spec)
}

func (a *Adapter) CleanupExpiredNamespaces(ctx context.Context, params worker.NamespaceCleanupParams) ([]worker.NamespaceCleanupResult, error) {
	return a.impl.CleanupExpiredNamespaces(ctx, params)
}

// Launch creates Kubernetes Job for run.
func (a *Adapter) Launch(ctx context.Context, spec worker.JobSpec) (worker.JobRef, error) {
	return a.impl.Launch(ctx, spec)
}

// Status returns current Kubernetes Job state.
func (a *Adapter) Status(ctx context.Context, ref worker.JobRef) (worker.JobState, error) {
	state, err := a.impl.Status(ctx, ref)
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
