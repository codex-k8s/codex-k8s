package project

import "context"

// Project is a persisted project catalog entry.
type Project struct {
	ID   string
	Slug string
	Name string
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
}

