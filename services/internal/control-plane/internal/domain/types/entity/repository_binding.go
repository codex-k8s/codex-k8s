package entity

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
