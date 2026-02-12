package platformtoken

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	PlatformGitHubTokens = entitytypes.PlatformGitHubTokens
	UpsertParams         = querytypes.PlatformGitHubTokensUpsertParams
)

// Repository stores singleton encrypted GitHub tokens used by platform runtime paths.
type Repository interface {
	// Get returns singleton token row.
	Get(ctx context.Context) (PlatformGitHubTokens, bool, error)
	// Upsert writes singleton token row.
	Upsert(ctx context.Context, params UpsertParams) (PlatformGitHubTokens, error)
}
