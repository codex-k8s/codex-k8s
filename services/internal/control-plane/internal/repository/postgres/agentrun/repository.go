package agentrun

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/agentrun/dbmodel"
)

var (
	//go:embed sql/create_pending_if_absent.sql
	queryCreatePendingIfAbsent string
	//go:embed sql/get_run_id_by_correlation_id.sql
	queryGetRunIDByCorrelationID string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/cancel_active_by_id.sql
	queryCancelActiveByID string
	//go:embed sql/list_run_ids_by_repository_issue.sql
	queryListRunIDsByRepositoryIssue string
	//go:embed sql/list_run_ids_by_repository_pull_request.sql
	queryListRunIDsByRepositoryPullRequest string
	//go:embed sql/list_recent_by_project.sql
	queryListRecentByProject string
	//go:embed sql/search_recent_by_project_issue_or_pull_request.sql
	querySearchRecentByProjectIssueOrPullRequest string
	//go:embed sql/release_slots_by_run_id.sql
	queryReleaseSlotsByRunID string
	//go:embed sql/set_wait_context.sql
	querySetWaitContext string
	//go:embed sql/clear_wait_context_if_matches.sql
	queryClearWaitContextIfMatches string
	//go:embed sql/upsert_run_agent_logs.sql
	queryUpsertRunAgentLogs string
	//go:embed sql/cleanup_run_agent_logs_finished_before.sql
	queryCleanupRunAgentLogsFinishedBefore string
)

// Repository stores agent runs in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL agent run repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreatePendingIfAbsent inserts pending run or returns an existing run id.
func (r *Repository) CreatePendingIfAbsent(ctx context.Context, params domainrepo.CreateParams) (domainrepo.CreateResult, error) {
	runID := uuid.NewString()
	var insertedRunID string
	err := r.db.QueryRow(
		ctx,
		queryCreatePendingIfAbsent,
		runID,
		params.CorrelationID,
		params.ProjectID,
		params.AgentID,
		[]byte(params.RunPayload),
		params.LearningMode,
	).Scan(&insertedRunID)
	if err == nil {
		return domainrepo.CreateResult{
			RunID:    insertedRunID,
			Inserted: true,
		}, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.CreateResult{}, fmt.Errorf("insert agent run: %w", err)
	}

	var existingRunID string
	if err := r.db.QueryRow(
		ctx,
		queryGetRunIDByCorrelationID,
		params.CorrelationID,
	).Scan(&existingRunID); err != nil {
		return domainrepo.CreateResult{}, fmt.Errorf("get existing agent run by correlation id: %w", err)
	}

	return domainrepo.CreateResult{
		RunID:    existingRunID,
		Inserted: false,
	}, nil
}

// GetByID returns one run by id.
func (r *Repository) GetByID(ctx context.Context, runID string) (domainrepo.Run, bool, error) {
	var row dbmodel.RunRow
	rows, err := r.db.Query(ctx, queryGetByID, runID)
	if err != nil {
		return domainrepo.Run{}, false, fmt.Errorf("query run by id: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RunRow])
	if err != nil {
		return domainrepo.Run{}, false, fmt.Errorf("collect run by id: %w", err)
	}
	if len(items) == 0 {
		return domainrepo.Run{}, false, nil
	}
	row = items[0]
	return runFromDBModel(row), true, nil
}

// CancelActiveByID marks one pending/running run as canceled and releases its slot lease when present.
func (r *Repository) CancelActiveByID(ctx context.Context, runID string) (bool, error) {
	trimmedRunID := strings.TrimSpace(runID)
	if trimmedRunID == "" {
		return false, fmt.Errorf("run_id is required")
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin cancel transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	res, err := tx.Exec(ctx, queryCancelActiveByID, trimmedRunID)
	if err != nil {
		return false, fmt.Errorf("cancel active run %s: %w", trimmedRunID, err)
	}
	if res.RowsAffected() == 0 {
		return false, nil
	}

	if _, err := tx.Exec(ctx, queryReleaseSlotsByRunID, trimmedRunID); err != nil {
		return false, fmt.Errorf("release slots for run %s: %w", trimmedRunID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit cancel transaction: %w", err)
	}

	return true, nil
}

// ListRecentByProject returns project runs ordered by newest first.
func (r *Repository) ListRecentByProject(ctx context.Context, projectID string, repositoryFullName string, limit int, offset int) ([]domainrepo.RunLookupItem, error) {
	trimmedProjectID := strings.TrimSpace(projectID)
	if trimmedProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = 50
	}
	if normalizedLimit > 200 {
		normalizedLimit = 200
	}
	normalizedOffset := offset
	if normalizedOffset < 0 {
		normalizedOffset = 0
	}

	rows, err := r.db.Query(ctx, queryListRecentByProject, trimmedProjectID, strings.TrimSpace(repositoryFullName), normalizedLimit, normalizedOffset)
	if err != nil {
		return nil, fmt.Errorf("list recent runs by project: %w", err)
	}
	runRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RunLookupRow])
	if err != nil {
		return nil, fmt.Errorf("collect recent runs by project: %w", err)
	}
	items := make([]domainrepo.RunLookupItem, 0, len(runRows))
	for _, row := range runRows {
		items = append(items, runLookupItemFromDBModel(row))
	}
	return items, nil
}

// SearchRecentByProjectIssueOrPullRequest returns project runs by issue/pr references ordered by newest first.
func (r *Repository) SearchRecentByProjectIssueOrPullRequest(ctx context.Context, projectID string, repositoryFullName string, issueNumber int64, pullRequestNumber int64, limit int) ([]domainrepo.RunLookupItem, error) {
	trimmedProjectID := strings.TrimSpace(projectID)
	if trimmedProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if issueNumber <= 0 && pullRequestNumber <= 0 {
		return nil, fmt.Errorf("issue_number or pull_request_number is required")
	}
	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = 50
	}
	if normalizedLimit > 200 {
		normalizedLimit = 200
	}

	rows, err := r.db.Query(
		ctx,
		querySearchRecentByProjectIssueOrPullRequest,
		trimmedProjectID,
		strings.TrimSpace(repositoryFullName),
		issueNumber,
		pullRequestNumber,
		normalizedLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("search recent runs by issue/pull request: %w", err)
	}
	runRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RunLookupRow])
	if err != nil {
		return nil, fmt.Errorf("collect recent runs by issue/pull request: %w", err)
	}
	items := make([]domainrepo.RunLookupItem, 0, len(runRows))
	for _, row := range runRows {
		items = append(items, runLookupItemFromDBModel(row))
	}
	return items, nil
}

// ListRunIDsByRepositoryIssue returns run ids for one repository/issue pair.
func (r *Repository) ListRunIDsByRepositoryIssue(ctx context.Context, repositoryFullName string, issueNumber int64, limit int) ([]string, error) {
	return r.listRunIDsByRepositoryReference(ctx, queryListRunIDsByRepositoryIssue, repositoryFullName, issueNumber, limit, "issue_number", "repository/issue")
}

// ListRunIDsByRepositoryPullRequest returns run ids for one repository/pull request pair.
func (r *Repository) ListRunIDsByRepositoryPullRequest(ctx context.Context, repositoryFullName string, prNumber int64, limit int) ([]string, error) {
	return r.listRunIDsByRepositoryReference(ctx, queryListRunIDsByRepositoryPullRequest, repositoryFullName, prNumber, limit, "pr_number", "repository/pull request")
}

// SetWaitContext updates typed wait linkage stored in agent_runs.
func (r *Repository) SetWaitContext(ctx context.Context, params domainrepo.SetWaitContextParams) (bool, error) {
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return false, fmt.Errorf("run_id is required")
	}

	res, err := r.db.Exec(
		ctx,
		querySetWaitContext,
		runID,
		string(params.WaitReason),
		string(params.WaitTargetKind),
		strings.TrimSpace(params.WaitTargetRef),
		timestamptzPtrOrNil(params.WaitDeadlineAt),
	)
	if err != nil {
		return false, fmt.Errorf("set run wait context: %w", err)
	}
	return res.RowsAffected() > 0, nil
}

// ClearWaitContextIfMatches clears wait linkage only when current wait still points to the expected target.
func (r *Repository) ClearWaitContextIfMatches(ctx context.Context, params domainrepo.ClearWaitContextParams) (bool, error) {
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return false, fmt.Errorf("run_id is required")
	}
	waitReason := strings.TrimSpace(string(params.WaitReason))
	if waitReason == "" {
		return false, fmt.Errorf("wait_reason is required")
	}
	waitTargetKind := strings.TrimSpace(string(params.WaitTargetKind))
	if waitTargetKind == "" {
		return false, fmt.Errorf("wait_target_kind is required")
	}
	waitTargetRef := strings.TrimSpace(params.WaitTargetRef)
	if waitTargetRef == "" {
		return false, fmt.Errorf("wait_target_ref is required")
	}

	res, err := r.db.Exec(
		ctx,
		queryClearWaitContextIfMatches,
		runID,
		waitReason,
		waitTargetKind,
		waitTargetRef,
	)
	if err != nil {
		return false, fmt.Errorf("clear run wait context if matches: %w", err)
	}
	return res.RowsAffected() > 0, nil
}

// UpsertRunAgentLogs stores latest agent execution logs for one run.
func (r *Repository) UpsertRunAgentLogs(ctx context.Context, runID string, logs json.RawMessage) error {
	trimmedRunID := strings.TrimSpace(runID)
	if trimmedRunID == "" {
		return fmt.Errorf("run_id is required")
	}

	payload := logs
	if len(payload) == 0 || !json.Valid(payload) {
		payload = json.RawMessage(`{}`)
	}

	if _, err := r.db.Exec(ctx, queryUpsertRunAgentLogs, trimmedRunID, []byte(payload)); err != nil {
		return fmt.Errorf("upsert run agent logs: %w", err)
	}
	return nil
}

// CleanupRunAgentLogsFinishedBefore clears logs for finished runs older than cutoff.
func (r *Repository) CleanupRunAgentLogsFinishedBefore(ctx context.Context, finishedBefore time.Time) (int64, error) {
	if finishedBefore.IsZero() {
		return 0, fmt.Errorf("finished_before is required")
	}

	res, err := r.db.Exec(ctx, queryCleanupRunAgentLogsFinishedBefore, finishedBefore.UTC())
	if err != nil {
		return 0, fmt.Errorf("cleanup run agent logs: %w", err)
	}
	return res.RowsAffected(), nil
}

func (r *Repository) listRunIDsByRepositoryReference(ctx context.Context, query string, repositoryFullName string, referenceNumber int64, limit int, referenceField string, operationLabel string) ([]string, error) {
	normalizedRepositoryFullName := strings.TrimSpace(repositoryFullName)
	if normalizedRepositoryFullName == "" {
		return nil, fmt.Errorf("repository_full_name is required")
	}
	if referenceNumber <= 0 {
		return nil, fmt.Errorf("%s must be positive", referenceField)
	}
	if limit <= 0 {
		limit = 200
	}

	rows, err := r.db.Query(ctx, query, normalizedRepositoryFullName, referenceNumber, limit)
	if err != nil {
		return nil, fmt.Errorf("list run ids by %s: %w", operationLabel, err)
	}
	runIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("collect run ids by %s: %w", operationLabel, err)
	}
	return runIDs, nil
}

func timestamptzPtrOrNil(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return value.UTC()
}
