package query

import entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"

// AgentListFilter keeps staff list filters for configured agents.
type AgentListFilter struct {
	Limit              int
	ProjectIDs         []string
	IncludeAllProjects bool
}

// AgentUpdateSettingsParams describes one optimistic settings update request.
type AgentUpdateSettingsParams struct {
	AgentID         string
	ExpectedVersion int
	UpdatedByUserID string
	Settings        entitytypes.AgentSettings
}
