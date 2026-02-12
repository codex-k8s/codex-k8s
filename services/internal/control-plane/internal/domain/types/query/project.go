package query

// ProjectUpsertParams defines inputs for creating or updating a project.
type ProjectUpsertParams struct {
	// ID is a project id to use for insert (server-generated in staff API).
	ID string
	// Slug is a stable project key (unique).
	Slug string
	// Name is a human-readable project name.
	Name string
	// SettingsJSON is a jsonb object stored in `projects.settings`.
	SettingsJSON []byte
}

// ProjectSettings stores project-level defaults in JSONB settings.
type ProjectSettings struct {
	LearningModeDefault bool `json:"learning_mode_default"`
}
