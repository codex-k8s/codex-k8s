package prompttemplate

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	KeyItem      = entitytypes.PromptTemplateKeyItem
	Version      = entitytypes.PromptTemplateVersion
	AuditEvent   = entitytypes.PromptTemplateAuditEvent
	SeedSyncItem = entitytypes.PromptTemplateSeedSyncItem
)

// Repository stores prompt template lifecycle state in PostgreSQL.
type Repository interface {
	ListKeys(ctx context.Context, filter querytypes.PromptTemplateKeyListFilter) ([]KeyItem, error)
	ListVersions(ctx context.Context, filter querytypes.PromptTemplateVersionListFilter) ([]Version, error)
	GetVersion(ctx context.Context, lookup querytypes.PromptTemplateVersionLookup) (Version, bool, error)
	GetActiveVersion(ctx context.Context, lookup querytypes.PromptTemplatePreviewLookup) (Version, bool, error)
	CreateVersion(ctx context.Context, params querytypes.PromptTemplateVersionCreateParams) (Version, error)
	ActivateVersion(ctx context.Context, params querytypes.PromptTemplateVersionActivateParams) (Version, error)
	CreateSeedIfMissing(ctx context.Context, params querytypes.PromptTemplateSeedCreateParams) (Version, bool, error)
	ListAuditEvents(ctx context.Context, filter querytypes.PromptTemplateAuditListFilter) ([]AuditEvent, error)
}

