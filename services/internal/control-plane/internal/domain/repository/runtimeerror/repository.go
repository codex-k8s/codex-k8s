package runtimeerror

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Item             = entitytypes.RuntimeError
	ListFilter       = querytypes.RuntimeErrorListFilter
	RecordParams     = querytypes.RuntimeErrorRecordParams
	MarkViewedParams = querytypes.RuntimeErrorMarkViewedParams
)

// Repository stores runtime error journal entries.
type Repository interface {
	// Insert appends one runtime error entry.
	Insert(ctx context.Context, params RecordParams) (Item, error)
	// ListAll returns runtime errors for platform admin scope.
	ListAll(ctx context.Context, filter ListFilter) ([]Item, error)
	// ListForUser returns runtime errors visible to projects where user is member.
	ListForUser(ctx context.Context, userID string, filter ListFilter) ([]Item, error)
	// GetByID returns one runtime error by id.
	GetByID(ctx context.Context, id string) (Item, bool, error)
	// MarkViewed marks runtime error as viewed by user and returns updated row.
	MarkViewed(ctx context.Context, params MarkViewedParams) (Item, bool, error)
}
