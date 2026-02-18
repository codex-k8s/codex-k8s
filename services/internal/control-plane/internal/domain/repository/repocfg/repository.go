package repocfg

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	RepositoryBinding                     = entitytypes.RepositoryBinding
	UpsertParams                          = querytypes.RepositoryBindingUpsertParams
	FindResult                            = querytypes.RepositoryBindingFindResult
	RepositoryBotParamsUpsertParams       = querytypes.RepositoryBotParamsUpsertParams
	RepositoryPreflightReportUpsertParams = querytypes.RepositoryPreflightReportUpsertParams
	RepositoryPreflightLockAcquireParams  = querytypes.RepositoryPreflightLockAcquireParams
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
	// FindByProviderOwnerName resolves configured binding for a provider repo slug.
	FindByProviderOwnerName(ctx context.Context, provider string, owner string, name string) (FindResult, bool, error)
	// GetTokenEncrypted returns encrypted token bytes for a repository binding.
	GetTokenEncrypted(ctx context.Context, repositoryID string) ([]byte, bool, error)
	// GetBotTokenEncrypted returns encrypted bot token bytes for a repository binding.
	GetBotTokenEncrypted(ctx context.Context, repositoryID string) ([]byte, bool, error)
	// UpsertBotParams updates bot token + params for a repository binding.
	UpsertBotParams(ctx context.Context, params RepositoryBotParamsUpsertParams) error
	// UpsertPreflightReport updates stored preflight report for a repository binding.
	UpsertPreflightReport(ctx context.Context, params RepositoryPreflightReportUpsertParams) error
	// AcquirePreflightLock acquires a short-lived lock to prevent concurrent preflight runs for one repo.
	AcquirePreflightLock(ctx context.Context, params RepositoryPreflightLockAcquireParams) (string, bool, error)
	// ReleasePreflightLock releases lock when token matches the owner token (best-effort).
	ReleasePreflightLock(ctx context.Context, repositoryID string, lockToken string) error
	// SetTokenEncryptedForAll updates token for all repository bindings.
	SetTokenEncryptedForAll(ctx context.Context, tokenEncrypted []byte) (int64, error)
}
