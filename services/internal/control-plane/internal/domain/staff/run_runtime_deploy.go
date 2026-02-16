package staff

import (
	"context"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
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
