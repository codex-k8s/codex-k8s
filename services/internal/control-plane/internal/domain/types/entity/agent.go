package entity

// Agent stores the subset of agents required by webhook run orchestration.
type Agent struct {
	ID        string
	AgentKey  string
	RoleKind  string
	ProjectID string
	Name      string
}
