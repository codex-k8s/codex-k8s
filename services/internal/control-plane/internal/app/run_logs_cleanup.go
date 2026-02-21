package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	agentcallbackdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/agentcallback"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
)

const (
	runHeavyFieldsCleanupInterval = time.Hour
	runHeavyFieldsCleanupTimeout  = 2 * time.Minute
)

func startRunHeavyFieldsCleanupLoop(ctx context.Context, callbacks *agentcallbackdomain.Service, runtimeDeploy *runtimedeploydomain.Service, logger *slog.Logger, retentionDays int) error {
	if callbacks == nil {
		return fmt.Errorf("agent callback service is required")
	}
	if runtimeDeploy == nil {
		return fmt.Errorf("runtime deploy service is required")
	}
	if retentionDays <= 0 {
		return fmt.Errorf("run heavy fields retention days must be > 0")
	}
	if logger == nil {
		logger = slog.Default()
	}

	retention := time.Duration(retentionDays) * 24 * time.Hour
	runCleanup := func() {
		now := time.Now().UTC()
		cleanupBefore := now.Add(-retention)

		cleanupCtx, cancel := context.WithTimeout(ctx, runHeavyFieldsCleanupTimeout)
		defer cancel()

		clearedRunLogs, err := callbacks.CleanupRunAgentLogs(cleanupCtx, cleanupBefore)
		if err != nil {
			logger.Error("run agent logs cleanup failed", "err", err, "cleanup_before", cleanupBefore.Format(time.RFC3339))
		}

		clearedSessionPayloads, err := callbacks.CleanupSessionPayloads(cleanupCtx, cleanupBefore)
		if err != nil {
			logger.Error("run session payload cleanup failed", "err", err, "cleanup_before", cleanupBefore.Format(time.RFC3339))
		}

		clearedTaskLogs, err := runtimeDeploy.CleanupTaskLogs(cleanupCtx, cleanupBefore)
		if err != nil {
			logger.Error("runtime deploy task logs cleanup failed", "err", err, "cleanup_before", cleanupBefore.Format(time.RFC3339))
		}

		if total := clearedRunLogs + clearedSessionPayloads + clearedTaskLogs; total > 0 {
			logger.Info(
				"run heavy fields cleanup completed",
				"cleared_run_logs", clearedRunLogs,
				"cleared_session_payloads", clearedSessionPayloads,
				"cleared_runtime_deploy_task_logs", clearedTaskLogs,
				"cleanup_before", cleanupBefore.Format(time.RFC3339),
			)
		}
	}

	runCleanup()

	go func() {
		ticker := time.NewTicker(runHeavyFieldsCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCleanup()
			}
		}
	}()

	return nil
}
