package repocfg

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	RepositoryBinding = entitytypes.RepositoryBinding
	UpsertParams      = querytypes.RepositoryBindingUpsertParams
	FindResult        = querytypes.RepositoryBindingFindResult
)

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
