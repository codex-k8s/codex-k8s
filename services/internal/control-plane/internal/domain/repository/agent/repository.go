package agent

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type Agent = entitytypes.Agent
type AgentSettings = entitytypes.AgentSettings

// Repository provides read access to configured agent profiles.
type Repository interface {
	// FindEffectiveByKey resolves active agent profile by key.
	// Project-scoped agent has priority over system profile.
	FindEffectiveByKey(ctx context.Context, projectID string, agentKey string) (Agent, bool, error)
	// List returns active agents visible for supplied project ids plus global defaults.
	List(ctx context.Context, filter querytypes.AgentListFilter) ([]Agent, error)
	// GetByID returns one agent by identifier.
	GetByID(ctx context.Context, agentID string) (Agent, bool, error)
	// UpdateSettings updates agent settings with optimistic concurrency.
	UpdateSettings(ctx context.Context, params querytypes.AgentUpdateSettingsParams) (Agent, bool, error)
}
