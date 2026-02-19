package runqueue

import (
	"context"

	querytypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/query"
)

type (
	ClaimParams  = querytypes.RunQueueClaimParams
	ClaimedRun   = querytypes.RunQueueClaimedRun
	RunningRun   = querytypes.RunQueueRunningRun
	FinishParams = querytypes.RunQueueFinishParams
)

// Repository provides queue-like operations over agent runs and slots.
type Repository interface {
	// ClaimNextPending atomically claims one pending run and leases a free slot when required by runtime profile.
	ClaimNextPending(ctx context.Context, params ClaimParams) (ClaimedRun, bool, error)
	// ListRunning returns active runs for reconciliation.
	ListRunning(ctx context.Context, limit int) ([]RunningRun, error)
	// FinishRun finalizes run status and releases slot lease when it exists.
	FinishRun(ctx context.Context, params FinishParams) (bool, error)
}
