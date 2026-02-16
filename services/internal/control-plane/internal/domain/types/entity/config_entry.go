package entity

// ConfigEntry is a persisted configuration entry (variable or secret).
//
// Secret values are never returned in plaintext; for secrets Value is always empty.
type ConfigEntry struct {
	ID string

	Scope string
	Kind  string

	ProjectID    string
	RepositoryID string

	Key string

	// Value is returned only for variables (kind=variable).
	Value string

	SyncTargets []string
	Mutability  string
	IsDangerous bool

	UpdatedAt string
}

