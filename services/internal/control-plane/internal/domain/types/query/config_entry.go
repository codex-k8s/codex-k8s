package query

// ConfigEntryUpsertParams defines inputs for configuration entry upsert.
type ConfigEntryUpsertParams struct {
	Scope string
	Kind  string

	ProjectID    string
	RepositoryID string

	Key string

	// ValuePlain is used for variables.
	ValuePlain string
	// ValueEncrypted is used for secrets.
	ValueEncrypted []byte

	SyncTargets []string
	Mutability  string
	IsDangerous bool

	CreatedByUserID string
	UpdatedByUserID string
}

type ConfigEntryListFilter struct {
	Scope        string
	ProjectID    string
	RepositoryID string
	Limit        int
}

