package projectmember

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

type Member = entitytypes.ProjectMember

// Repository stores and loads project memberships.
type Repository interface {
	// List returns project members.
	List(ctx context.Context, projectID string, limit int) ([]Member, error)
	// Upsert sets role for a user in a project.
	Upsert(ctx context.Context, projectID string, userID string, role string) error
	// Delete removes a user from a project.
	Delete(ctx context.Context, projectID string, userID string) error
	// GetRole returns membership role for a user in a project.
	GetRole(ctx context.Context, projectID string, userID string) (role string, ok bool, err error)

	// SetLearningModeOverride sets per-member learning mode override.
	// When enabled is nil, the override is removed and project default is used.
	SetLearningModeOverride(ctx context.Context, projectID string, userID string, enabled *bool) error

	// GetLearningModeOverride returns per-member learning mode override.
	// When ok is false, there is no membership record.
	// When override is nil, the override is not set (inherit project default).
	GetLearningModeOverride(ctx context.Context, projectID string, userID string) (override *bool, ok bool, err error)
}
