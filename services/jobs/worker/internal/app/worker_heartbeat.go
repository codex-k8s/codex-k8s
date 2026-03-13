package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	workerinstancerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/workerinstance"
)

type workerHeartbeatParams struct {
	WorkerID  string
	Namespace string
	PodName   string
	StartedAt time.Time
	TTL       time.Duration
	Now       time.Time
}

type workerHeartbeatLoopParams struct {
	WorkerID  string
	Namespace string
	PodName   string
	StartedAt time.Time
	Interval  time.Duration
	TTL       time.Duration
}

func workerHeartbeat(ctx context.Context, repo workerinstancerepo.Repository, params workerHeartbeatParams) error {
	now := params.Now.UTC()
	return repo.Heartbeat(ctx, workerinstancerepo.HeartbeatParams{
		WorkerID:    params.WorkerID,
		Namespace:   params.Namespace,
		PodName:     params.PodName,
		StartedAt:   params.StartedAt.UTC(),
		HeartbeatAt: now,
		ExpiresAt:   now.Add(params.TTL),
	})
}

func runWorkerHeartbeatLoop(ctx context.Context, logger *slog.Logger, repo workerinstancerepo.Repository, params workerHeartbeatLoopParams) {
	ticker := time.NewTicker(params.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case tickAt := <-ticker.C:
			heartbeatCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := workerHeartbeat(heartbeatCtx, repo, workerHeartbeatParams{
				WorkerID:  params.WorkerID,
				Namespace: params.Namespace,
				PodName:   params.PodName,
				StartedAt: params.StartedAt,
				TTL:       params.TTL,
				Now:       tickAt.UTC(),
			})
			cancel()
			if err != nil {
				logger.Warn("worker heartbeat failed", "worker_id", params.WorkerID, "err", err)
			}
		}
	}
}

func markWorkerStopped(ctx context.Context, repo workerinstancerepo.Repository, workerID string, stoppedAt time.Time) error {
	if err := repo.MarkStopped(ctx, workerinstancerepo.StopParams{
		WorkerID:  workerID,
		StoppedAt: stoppedAt.UTC(),
	}); err != nil {
		return fmt.Errorf("mark worker stopped: %w", err)
	}
	return nil
}
