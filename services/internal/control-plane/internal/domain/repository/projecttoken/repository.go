package projecttoken

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	ProjectGitHubTokens = entitytypes.ProjectGitHubTokens
	UpsertParams        = querytypes.ProjectGitHubTokensUpsertParams
)

// Repository persists project-scoped GitHub tokens.
type Repository interface {
	GetByProjectID(ctx context.Context, projectID string) (ProjectGitHubTokens, bool, error)
	GetEncryptedByProjectID(ctx context.Context, projectID string) (platformToken []byte, botToken []byte, botUsername string, botEmail string, ok bool, err error)
	Upsert(ctx context.Context, params UpsertParams) error
	DeleteByProjectID(ctx context.Context, projectID string) error
}

