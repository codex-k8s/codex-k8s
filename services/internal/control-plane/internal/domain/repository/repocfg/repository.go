package repocfg

import "context"

// RepositoryBinding represents a repository attached to a project.
//
// This record contains an encrypted access token (stored in DB) and per-repository configuration,
// such as `services.yaml` path override.
type RepositoryBinding struct {
	ID string

	ProjectID string

	// Provider is a repository hosting provider id (e.g. "github").
	Provider string

	// ExternalID is a provider-specific repository id (e.g. GitHub repository numeric id).
	ExternalID int64

	// Owner is a repository owner/namespace (e.g. "codex-k8s").
	Owner string

	// Name is a repository short name (e.g. "codex-k8s").
	Name string

	// ServicesYAMLPath is a path to services.yaml within the repository.
	ServicesYAMLPath string
}

// UpsertParams defines inputs for repository binding upsert.
type UpsertParams struct {
	ProjectID  string
	Provider   string
	ExternalID int64
	Owner      string
	Name       string

	// TokenEncrypted stores encrypted token bytes (nonce||ciphertext) for DB BYTEA column.
	TokenEncrypted []byte

	ServicesYAMLPath string
}

// FindResult is a lookup result for webhook->project binding resolution.
type FindResult struct {
	ProjectID        string
	RepositoryID     string
	ServicesYAMLPath string
}

// Repository persists project repository bindings.
type Repository interface {
	// ListForProject returns repository bindings for a project.
	ListForProject(ctx context.Context, projectID string, limit int) ([]RepositoryBinding, error)
	// GetByID returns one repository binding by id.
	GetByID(ctx context.Context, repositoryID string) (RepositoryBinding, bool, error)
	// Upsert creates/updates a binding (unique by provider+external_id).
	Upsert(ctx context.Context, params UpsertParams) (RepositoryBinding, error)
	// Delete removes a binding by id within a project.
	Delete(ctx context.Context, projectID string, repositoryID string) error
	// FindByProviderExternalID resolves configured binding for a provider repo id.
	FindByProviderExternalID(ctx context.Context, provider string, externalID int64) (FindResult, bool, error)
	// GetTokenEncrypted returns encrypted token bytes for a repository binding.
	GetTokenEncrypted(ctx context.Context, repositoryID string) ([]byte, bool, error)
}
