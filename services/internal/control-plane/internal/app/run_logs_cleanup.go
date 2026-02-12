package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	agentcallbackdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/agentcallback"
)

const (
	runAgentLogsCleanupInterval = time.Hour
	runAgentLogsCleanupTimeout  = 30 * time.Second
)

func startRunAgentLogsCleanupLoop(ctx context.Context, callbacks *agentcallbackdomain.Service, logger *slog.Logger, retentionDays int) error {
	if callbacks == nil {
		return fmt.Errorf("agent callback service is required")
	}
	if retentionDays <= 0 {
		return fmt.Errorf("run agent logs retention days must be > 0")
	}
	if logger == nil {
		logger = slog.Default()
	}

	retention := time.Duration(retentionDays) * 24 * time.Hour
	runCleanup := func() {
		now := time.Now().UTC()
		cleanupBefore := now.Add(-retention)

		cleanupCtx, cancel := context.WithTimeout(ctx, runAgentLogsCleanupTimeout)
		defer cancel()

		cleared, err := callbacks.CleanupRunAgentLogs(cleanupCtx, cleanupBefore)
		if err != nil {
			logger.Error("run agent logs cleanup failed", "err", err, "cleanup_before", cleanupBefore.Format(time.RFC3339))
			return
		}
		if cleared > 0 {
			logger.Info("run agent logs cleanup completed", "cleared_runs", cleared, "cleanup_before", cleanupBefore.Format(time.RFC3339))
		}
	}

	runCleanup()

	go func() {
		ticker := time.NewTicker(runAgentLogsCleanupInterval)
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
