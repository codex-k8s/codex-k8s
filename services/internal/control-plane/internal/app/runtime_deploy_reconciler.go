package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	runtimedeploydomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/runtimedeploy"
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
	workersPerPod int,
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
	if workersPerPod <= 0 {
		return fmt.Errorf("runtime deploy workers per pod must be > 0")
	}

	runOnce := func(leaseOwner string) {
		for {
			processed, err := reconciler.ReconcileNext(ctx, leaseOwner, leaseTTL)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.Error("runtime deploy reconcile tick failed", "worker_id", leaseOwner, "err", err)
				return
			}
			if !processed {
				return
			}
		}
	}

	for idx := 0; idx < workersPerPod; idx++ {
		leaseOwner := workerID
		if workersPerPod > 1 {
			leaseOwner = workerID + "-w" + strconv.Itoa(idx+1)
		}

		runOnce(leaseOwner)

		go func(owner string) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					runOnce(owner)
				}
			}
		}(leaseOwner)
	}

	logger.Info(
		"runtime deploy reconciler loop started",
		"worker_id",
		workerID,
		"interval",
		interval.String(),
		"lease_ttl",
		leaseTTL.String(),
		"workers_per_pod",
		workersPerPod,
	)
	return nil
}

var _ runtimeDeployReconciler = (*runtimedeploydomain.Service)(nil)
