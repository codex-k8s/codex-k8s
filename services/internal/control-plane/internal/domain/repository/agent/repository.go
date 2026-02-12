package agent

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

type Agent = entitytypes.Agent

// Repository provides read access to configured agent profiles.
type Repository interface {
	// FindEffectiveByKey resolves active agent profile by key.
	// Project-scoped agent has priority over system profile.
	FindEffectiveByKey(ctx context.Context, projectID string, agentKey string) (Agent, bool, error)
}
