package entity

// AgentSettings stores runtime/prompt behavior controlled from staff UI.
type AgentSettings struct {
	RuntimeMode       string
	TimeoutSeconds    int
	MaxRetryCount     int
	PromptLocale      string
	ApprovalsRequired bool
}

// Agent stores agent profile and staff-managed settings.
type Agent struct {
	ID              string
	AgentKey        string
	RoleKind        string
	ProjectID       string
	Name            string
	IsActive        bool
	Settings        AgentSettings
	SettingsVersion int
}
