package workerinstance

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	domainrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/workerinstance"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/upsert_heartbeat.sql
	queryUpsertHeartbeat string
	//go:embed sql/mark_stopped.sql
	queryMarkStopped string
)

// Repository persists worker liveness records in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs worker liveness repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Heartbeat registers or refreshes one worker instance row.
func (r *Repository) Heartbeat(ctx context.Context, params domainrepo.HeartbeatParams) error {
	workerID := strings.TrimSpace(params.WorkerID)
	if workerID == "" {
		return fmt.Errorf("worker heartbeat: worker_id is required")
	}

	if _, err := r.db.Exec(
		ctx,
		queryUpsertHeartbeat,
		workerID,
		strings.TrimSpace(params.Namespace),
		strings.TrimSpace(params.PodName),
		params.StartedAt.UTC(),
		params.HeartbeatAt.UTC(),
		params.ExpiresAt.UTC(),
	); err != nil {
		return fmt.Errorf("worker heartbeat: %w", err)
	}
	return nil
}

// MarkStopped marks worker instance as stopped for faster lease recovery.
func (r *Repository) MarkStopped(ctx context.Context, params domainrepo.StopParams) error {
	workerID := strings.TrimSpace(params.WorkerID)
	if workerID == "" {
		return nil
	}

	if _, err := r.db.Exec(ctx, queryMarkStopped, workerID, params.StoppedAt.UTC()); err != nil {
		return fmt.Errorf("mark worker stopped: %w", err)
	}
	return nil
}
