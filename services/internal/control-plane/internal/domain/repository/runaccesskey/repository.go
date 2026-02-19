package runaccesskey

import (
	"context"
	"time"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Run          = entitytypes.RunAccessKey
	UpsertParams = querytypes.RunAccessKeyUpsertParams
)

// Repository stores run-scoped access keys used for controlled OAuth bypass.
type Repository interface {
	// GetByRunID returns one run access key row by run id.
	GetByRunID(ctx context.Context, runID string) (Run, bool, error)
	// Upsert creates or updates run access key row.
	Upsert(ctx context.Context, params UpsertParams) (Run, error)
	// Revoke marks run access key as revoked.
	Revoke(ctx context.Context, runID string, revokedAt time.Time, updatedAt time.Time) (Run, bool, error)
	// TouchLastUsed updates last_used_at timestamp after successful authorization.
	TouchLastUsed(ctx context.Context, runID string, usedAt time.Time) error
}
