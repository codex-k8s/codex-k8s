package runtimedeploy

import (
	"context"
	"encoding/json"
	"strings"

	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const maxRuntimeDeployTaskLogMessageLength = 4000

func (s *Service) appendTaskLogBestEffort(ctx context.Context, runID string, stage string, level string, message string) {
	if s == nil || s.tasks == nil {
		return
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	if len(message) > maxRuntimeDeployTaskLogMessageLength {
		message = message[:maxRuntimeDeployTaskLogMessageLength]
	}
	if err := s.tasks.AppendLog(ctx, runtimedeploytaskrepo.AppendLogParams{
		RunID:    runID,
		Stage:    strings.TrimSpace(stage),
		Level:    strings.TrimSpace(level),
		Message:  message,
		MaxLines: 300,
	}); err != nil {
		s.logger.Warn("append runtime deploy task log failed", "run_id", runID, "stage", stage, "level", level, "err", err)
	}
	if strings.EqualFold(strings.TrimSpace(level), "error") && s.runtimeErr != nil {
		source := "control-plane.runtime-deploy"
		if stageToken := strings.TrimSpace(stage); stageToken != "" {
			source += "." + stageToken
		}
		details, _ := json.Marshal(map[string]string{
			"channel": "runtime_deploy_task_log",
			"stage":   strings.TrimSpace(stage),
		})
		s.runtimeErr.RecordBestEffort(ctx, querytypes.RuntimeErrorRecordParams{
			Source:      source,
			Level:       "error",
			Message:     message,
			DetailsJSON: details,
			RunID:       runID,
		})
	}
}
