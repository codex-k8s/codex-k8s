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
}
