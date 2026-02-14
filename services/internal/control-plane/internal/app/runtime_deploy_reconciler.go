package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
)

type runtimeDeployReconciler interface {
	ReconcileNext(ctx context.Context, leaseOwner string, leaseTTL time.Duration) (bool, error)
}

func startRuntimeDeployReconcilerLoop(
	ctx context.Context,
	reconciler runtimeDeployReconciler,
	logger *slog.Logger,
	workerID string,
	interval time.Duration,
	leaseTTL time.Duration,
) error {
	if reconciler == nil {
		return fmt.Errorf("runtime deploy reconciler is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return fmt.Errorf("runtime deploy worker id is required")
	}
	if interval <= 0 {
		return fmt.Errorf("runtime deploy reconcile interval must be > 0")
	}
	if leaseTTL <= 0 {
		return fmt.Errorf("runtime deploy lease ttl must be > 0")
	}

	runOnce := func() {
		for {
			processed, err := reconciler.ReconcileNext(ctx, workerID, leaseTTL)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.Error("runtime deploy reconcile tick failed", "worker_id", workerID, "err", err)
				return
			}
			if !processed {
				return
			}
		}
	}

	runOnce()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runOnce()
			}
		}
	}()

	logger.Info("runtime deploy reconciler loop started", "worker_id", workerID, "interval", interval.String(), "lease_ttl", leaseTTL.String())
	return nil
}

var _ runtimeDeployReconciler = (*runtimedeploydomain.Service)(nil)
