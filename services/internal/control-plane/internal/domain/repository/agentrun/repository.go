package agentrun

import (
	"context"

	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type Run = entity.AgentRun
type CreateParams = query.AgentRunCreateParams
type CreateResult = query.AgentRunCreateResult
type RunLookupItem = query.AgentRunLookupItem

// Repository persists and queries agent run records.
type Repository interface {
	// CreatePendingIfAbsent inserts a pending run unless it already exists.
	CreatePendingIfAbsent(ctx context.Context, params CreateParams) (CreateResult, error)
	// GetByID returns one run by id.
	GetByID(ctx context.Context, runID string) (Run, bool, error)
	// ListRecentByProject returns project runs ordered by newest first.
	ListRecentByProject(ctx context.Context, projectID string, repositoryFullName string, limit int, offset int) ([]RunLookupItem, error)
	// SearchRecentByProjectIssueOrPullRequest returns project runs by issue/pr references ordered by newest first.
	SearchRecentByProjectIssueOrPullRequest(ctx context.Context, projectID string, repositoryFullName string, issueNumber int64, pullRequestNumber int64, limit int) ([]RunLookupItem, error)
	// ListRunIDsByRepositoryIssue returns run ids for one repository/issue pair.
	ListRunIDsByRepositoryIssue(ctx context.Context, repositoryFullName string, issueNumber int64, limit int) ([]string, error)
	// ListRunIDsByRepositoryPullRequest returns run ids for one repository/pull request pair.
	ListRunIDsByRepositoryPullRequest(ctx context.Context, repositoryFullName string, prNumber int64, limit int) ([]string, error)
}
