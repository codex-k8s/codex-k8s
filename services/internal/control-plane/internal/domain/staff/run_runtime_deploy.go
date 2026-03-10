package staff

import (
	"context"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
)

const defaultRuntimeDeployTaskLimit = 200

// ListRuntimeDeployTasks returns runtime deploy task list (platform admin only).
func (s *Service) ListRuntimeDeployTasks(ctx context.Context, principal Principal, limit int, status string, targetEnv string) ([]runtimedeploytaskrepo.Task, error) {
	if !principal.IsPlatformAdmin {
		return nil, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.tasks == nil {
		return nil, errs.Validation{Field: "runtime_deploy", Msg: "task repository is not configured"}
	}
	if limit <= 0 {
		limit = defaultRuntimeDeployTaskLimit
	}
	items, err := s.tasks.ListRecent(ctx, runtimedeploytaskrepo.ListFilter{
		Limit:     limit,
		Status:    strings.TrimSpace(status),
		TargetEnv: strings.TrimSpace(targetEnv),
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetRuntimeDeployTask returns one runtime deploy task by run id (platform admin only).
func (s *Service) GetRuntimeDeployTask(ctx context.Context, principal Principal, runID string) (runtimedeploytaskrepo.Task, error) {
	if !principal.IsPlatformAdmin {
		return runtimedeploytaskrepo.Task{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.tasks == nil {
		return runtimedeploytaskrepo.Task{}, errs.Validation{Field: "runtime_deploy", Msg: "task repository is not configured"}
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return runtimedeploytaskrepo.Task{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}
	item, ok, err := s.tasks.GetByRunID(ctx, runID)
	if err != nil {
		return runtimedeploytaskrepo.Task{}, err
	}
	if !ok {
		return runtimedeploytaskrepo.Task{}, errs.Validation{Field: "run_id", Msg: "not found"}
	}
	return item, nil
}

// RequestRuntimeDeployTaskAction applies one cancel/stop control action to runtime deploy task.
func (s *Service) RequestRuntimeDeployTaskAction(
	ctx context.Context,
	principal Principal,
	runID string,
	action runtimedeploydomain.TaskAction,
	reason string,
) (runtimedeploydomain.TaskActionResult, error) {
	if !principal.IsPlatformAdmin {
		return runtimedeploydomain.TaskActionResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.runtimeDeploy == nil {
		return runtimedeploydomain.TaskActionResult{}, errs.Validation{Field: "runtime_deploy", Msg: "runtime deploy service is not configured"}
	}
	return s.runtimeDeploy.RequestTaskAction(ctx, runtimedeploydomain.TaskActionParams{
		RunID:  strings.TrimSpace(runID),
		Action: action,
		Reason: strings.TrimSpace(reason),
		Actor: runtimedeploydomain.TaskActionActor{
			UserID:      principal.UserID,
			Email:       principal.Email,
			GitHubLogin: principal.GitHubLogin,
		},
	})
}
