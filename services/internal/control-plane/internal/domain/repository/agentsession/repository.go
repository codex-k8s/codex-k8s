package agentsession

import (
	"context"
	"time"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Session            = entitytypes.AgentSession
	UpsertParams       = querytypes.AgentSessionUpsertParams
	SetWaitStateParams = querytypes.AgentSessionSetWaitStateParams
)

// Repository persists resumable codex-cli sessions for agent runs.
type Repository interface {
	// Upsert stores or updates run session snapshot by run_id.
	Upsert(ctx context.Context, params UpsertParams) error
	// SetWaitStateByRunID updates wait-state and timeout guard fields for the latest run session.
	SetWaitStateByRunID(ctx context.Context, params SetWaitStateParams) (bool, error)
	// GetByRunID returns latest session snapshot for one run id.
	GetByRunID(ctx context.Context, runID string) (Session, bool, error)
	// GetLatestByRepositoryBranchAndAgent returns latest snapshot by repository + branch + agent key.
	GetLatestByRepositoryBranchAndAgent(ctx context.Context, repositoryFullName string, branchName string, agentKey string) (Session, bool, error)
	// CleanupSessionPayloadsFinishedBefore clears heavy JSON payloads for finished runs older than cutoff.
	CleanupSessionPayloadsFinishedBefore(ctx context.Context, finishedBefore time.Time) (int64, error)
}
