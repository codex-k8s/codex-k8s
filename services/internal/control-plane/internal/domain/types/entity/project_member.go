package entity

// ProjectMember is a project membership record (joined with user email for listing).
type ProjectMember struct {
	ProjectID string
	UserID    string
	Email     string
	Role      string

	// LearningModeOverride is a tri-state override:
	// nil => inherit project default, true/false => explicit override.
	LearningModeOverride *bool
}
