package workerinstance

import (
	"context"

	querytypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/query"
)

type (
	HeartbeatParams = querytypes.WorkerInstanceHeartbeatParams
	StopParams      = querytypes.WorkerInstanceStopParams
)

// Repository persists worker liveness state in PostgreSQL.
type Repository interface {
	// Heartbeat registers or refreshes one worker instance liveness record.
	Heartbeat(ctx context.Context, params HeartbeatParams) error
	// MarkStopped marks one worker instance as gracefully stopped.
	MarkStopped(ctx context.Context, params StopParams) error
}
