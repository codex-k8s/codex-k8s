package main

import (
	"context"
	"fmt"

	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
)

// noopRuntimeDeployTaskRepository is used by one-shot runtime deploy CLI.
// ApplyNow path only appends logs; queue methods must never be called here.
type noopRuntimeDeployTaskRepository struct{}

func (noopRuntimeDeployTaskRepository) UpsertDesired(_ context.Context, _ runtimedeploytaskrepo.UpsertDesiredParams) (runtimedeploytaskrepo.Task, error) {
	return runtimedeploytaskrepo.Task{}, fmt.Errorf("runtime deploy queue is not available in one-shot mode")
}

func (noopRuntimeDeployTaskRepository) GetByRunID(_ context.Context, _ string) (runtimedeploytaskrepo.Task, bool, error) {
	return runtimedeploytaskrepo.Task{}, false, fmt.Errorf("runtime deploy queue is not available in one-shot mode")
}

func (noopRuntimeDeployTaskRepository) ClaimNext(_ context.Context, _ runtimedeploytaskrepo.ClaimParams) (runtimedeploytaskrepo.Task, bool, error) {
	return runtimedeploytaskrepo.Task{}, false, fmt.Errorf("runtime deploy queue is not available in one-shot mode")
}

func (noopRuntimeDeployTaskRepository) MarkSucceeded(_ context.Context, _ runtimedeploytaskrepo.MarkSucceededParams) (bool, error) {
	return false, fmt.Errorf("runtime deploy queue is not available in one-shot mode")
}

func (noopRuntimeDeployTaskRepository) MarkFailed(_ context.Context, _ runtimedeploytaskrepo.MarkFailedParams) (bool, error) {
	return false, fmt.Errorf("runtime deploy queue is not available in one-shot mode")
}

func (noopRuntimeDeployTaskRepository) RenewLease(_ context.Context, _ runtimedeploytaskrepo.RenewLeaseParams) (bool, error) {
	return false, fmt.Errorf("runtime deploy queue is not available in one-shot mode")
}

func (noopRuntimeDeployTaskRepository) ListRecent(_ context.Context, _ runtimedeploytaskrepo.ListFilter) ([]runtimedeploytaskrepo.Task, error) {
	return nil, fmt.Errorf("runtime deploy queue is not available in one-shot mode")
}

func (noopRuntimeDeployTaskRepository) AppendLog(_ context.Context, _ runtimedeploytaskrepo.AppendLogParams) error {
	return nil
}
