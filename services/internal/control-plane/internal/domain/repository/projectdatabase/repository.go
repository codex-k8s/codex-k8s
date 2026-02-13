package projectdatabase

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Item         = entitytypes.ProjectDatabase
	UpsertParams = querytypes.ProjectDatabaseUpsertParams
)

// Repository stores ownership mapping for project-managed databases.
type Repository interface {
	// GetByDatabaseName returns one ownership row by global database name.
	GetByDatabaseName(ctx context.Context, databaseName string) (Item, bool, error)
	// Upsert creates or updates ownership mapping.
	Upsert(ctx context.Context, params UpsertParams) (Item, error)
	// DeleteByDatabaseName removes ownership mapping by global database name.
	DeleteByDatabaseName(ctx context.Context, databaseName string) (bool, error)
}
