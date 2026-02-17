package configentry

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	ConfigEntry  = entitytypes.ConfigEntry
	UpsertParams = querytypes.ConfigEntryUpsertParams
	ListFilter   = querytypes.ConfigEntryListFilter
)

// Repository persists config entries.
type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]ConfigEntry, error)
	GetByID(ctx context.Context, id string) (ConfigEntry, bool, error)
	Exists(ctx context.Context, scope string, projectID string, repositoryID string, key string) (bool, error)
	Upsert(ctx context.Context, params UpsertParams) (ConfigEntry, error)
	Delete(ctx context.Context, id string) error
}
