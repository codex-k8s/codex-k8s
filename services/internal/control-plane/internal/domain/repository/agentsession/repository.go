package agentsession

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Session      = entitytypes.AgentSession
	UpsertParams = querytypes.AgentSessionUpsertParams
)

// Repository persists resumable codex-cli sessions for agent runs.
type Repository interface {
	// Upsert stores or updates run session snapshot by run_id.
	Upsert(ctx context.Context, params UpsertParams) error
	// GetLatestByRepositoryBranchAndAgent returns latest snapshot by repository + branch + agent key.
	GetLatestByRepositoryBranchAndAgent(ctx context.Context, repositoryFullName string, branchName string, agentKey string) (Session, bool, error)
}
