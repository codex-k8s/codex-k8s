package query

// RepositoryBindingUpsertParams defines inputs for repository binding upsert.
type RepositoryBindingUpsertParams struct {
	ProjectID  string
	Alias      string
	Role       string
	DefaultRef string
	Provider   string
	ExternalID int64
	Owner      string
	Name       string

	// TokenEncrypted stores encrypted token bytes (nonce||ciphertext) for DB BYTEA column.
	TokenEncrypted []byte

	ServicesYAMLPath string
	DocsRootPath     string
}

// RepositoryBindingFindResult is a lookup result for webhook->project binding resolution.
type RepositoryBindingFindResult struct {
	ProjectID        string
	RepositoryID     string
	ServicesYAMLPath string
	DefaultRef       string
}
