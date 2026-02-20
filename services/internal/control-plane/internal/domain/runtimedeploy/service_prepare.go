package runtimedeploy

import (
	"context"
	"fmt"
	"strings"
	"time"

	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

// PrepareRunEnvironment stores desired deploy state and waits for reconciler completion.
func (s *Service) PrepareRunEnvironment(ctx context.Context, params PrepareParams) (PrepareResult, error) {
	params = normalizePrepareParams(params)
	if strings.TrimSpace(params.RunID) == "" {
		return PrepareResult{}, fmt.Errorf("run_id is required")
	}

	task, err := s.tasks.UpsertDesired(ctx, runtimedeploytaskrepo.UpsertDesiredParams{
		RunID:              params.RunID,
		RuntimeMode:        params.RuntimeMode,
		Namespace:          params.Namespace,
		TargetEnv:          params.TargetEnv,
		SlotNo:             params.SlotNo,
		RepositoryFullName: params.RepositoryFullName,
		ServicesYAMLPath:   params.ServicesYAMLPath,
		BuildRef:           params.BuildRef,
		DeployOnly:         params.DeployOnly,
	})
	if err != nil {
		return PrepareResult{}, fmt.Errorf("upsert runtime deploy task: %w", err)
	}
	if task.Status == entitytypes.RuntimeDeployTaskStatusSucceeded {
		return taskToPrepareResult(task), nil
	}

	return s.waitForTaskResult(ctx, params.RunID)
}

func (s *Service) waitForTaskResult(ctx context.Context, runID string) (PrepareResult, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return PrepareResult{}, fmt.Errorf("run_id is required")
	}

	for {
		task, ok, err := s.tasks.GetByRunID(ctx, runID)
		if err != nil {
			return PrepareResult{}, fmt.Errorf("load runtime deploy task run_id=%s: %w", runID, err)
		}
		if !ok {
			return PrepareResult{}, fmt.Errorf("runtime deploy task for run_id=%s not found", runID)
		}

		switch task.Status {
		case entitytypes.RuntimeDeployTaskStatusSucceeded:
			return taskToPrepareResult(task), nil
		case entitytypes.RuntimeDeployTaskStatusFailed:
			if strings.TrimSpace(task.LastError) == "" {
				return PrepareResult{}, fmt.Errorf("runtime deploy task failed for run_id=%s", runID)
			}
			return PrepareResult{}, fmt.Errorf("runtime deploy task failed for run_id=%s: %s", runID, task.LastError)
		case entitytypes.RuntimeDeployTaskStatusCanceled:
			return PrepareResult{}, TaskCanceledError{
				RunID:  runID,
				Reason: strings.TrimSpace(task.LastError),
			}
		}

		timer := time.NewTimer(s.cfg.WaitPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return PrepareResult{}, ctx.Err()
		case <-timer.C:
		}
	}
}

func normalizePrepareParams(params PrepareParams) PrepareParams {
	params.RunID = strings.TrimSpace(params.RunID)
	params.RuntimeMode = strings.TrimSpace(params.RuntimeMode)
	if params.RuntimeMode == "" {
		params.RuntimeMode = "full-env"
	}
	params.Namespace = strings.TrimSpace(params.Namespace)
	params.TargetEnv = strings.TrimSpace(params.TargetEnv)
	if params.TargetEnv == "" {
		params.TargetEnv = "ai"
	}
	if params.SlotNo < 0 {
		params.SlotNo = 0
	}
	params.RepositoryFullName = strings.TrimSpace(params.RepositoryFullName)
	params.ServicesYAMLPath = strings.TrimSpace(params.ServicesYAMLPath)
	params.BuildRef = strings.TrimSpace(params.BuildRef)
	return params
}

func taskToPrepareResult(task runtimedeploytaskrepo.Task) PrepareResult {
	namespace := strings.TrimSpace(task.ResultNamespace)
	if namespace == "" {
		namespace = strings.TrimSpace(task.Namespace)
	}
	targetEnv := strings.TrimSpace(task.ResultTargetEnv)
	if targetEnv == "" {
		targetEnv = strings.TrimSpace(task.TargetEnv)
	}
	if targetEnv == "" {
		targetEnv = "ai"
	}
	return PrepareResult{
		Namespace: namespace,
		TargetEnv: targetEnv,
	}
}
