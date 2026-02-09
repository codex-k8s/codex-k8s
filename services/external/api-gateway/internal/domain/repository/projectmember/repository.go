package projectmember

import "context"

// Member is a project membership record (joined with user email for listing).
type Member struct {
	ProjectID string
	UserID    string
	Email     string
	Role      string
}

// Repository stores and loads project memberships.
type Repository interface {
	// List returns project members.
	List(ctx context.Context, projectID string, limit int) ([]Member, error)
	// Upsert sets role for a user in a project.
	Upsert(ctx context.Context, projectID string, userID string, role string) error
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
