package project

import "context"

// Project is a persisted project catalog entry.
type Project struct {
	ID   string
	Slug string
	Name string
}

// UpsertParams defines inputs for creating or updating a project.
type UpsertParams struct {
	// ID is a project id to use for insert (server-generated in staff API).
	ID string
	// Slug is a stable project key (unique).
	Slug string
	// Name is a human-readable project name.
	Name string
	// SettingsJSON is a jsonb object stored in `projects.settings`.
	SettingsJSON []byte
}

// ProjectWithRole extends Project with an effective role for a user.
type ProjectWithRole struct {
	Project
	Role string
}

// Repository stores and loads projects.
type Repository interface {
	// ListAll returns all projects (platform admins).
	ListAll(ctx context.Context, limit int) ([]Project, error)
	// ListForUser returns projects visible to a user with their role.
	ListForUser(ctx context.Context, userID string, limit int) ([]ProjectWithRole, error)

	// Upsert creates/updates a project by slug.
	Upsert(ctx context.Context, params UpsertParams) (Project, error)

	// GetByID returns a project by id.
	GetByID(ctx context.Context, projectID string) (Project, bool, error)
	// DeleteByID deletes a project by id.
	DeleteByID(ctx context.Context, projectID string) error

	// GetLearningModeDefault returns effective project default learning-mode flag.
	GetLearningModeDefault(ctx context.Context, projectID string) (enabled bool, ok bool, err error)
}
