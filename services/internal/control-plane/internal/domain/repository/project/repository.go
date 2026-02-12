package project

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Project         = entitytypes.Project
	UpsertParams    = querytypes.ProjectUpsertParams
	ProjectWithRole = entitytypes.ProjectWithRole
)

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
