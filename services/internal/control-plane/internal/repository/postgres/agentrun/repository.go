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
	//go:embed sql/list_run_ids_by_repository_issue.sql
	queryListRunIDsByRepositoryIssue string
	//go:embed sql/list_run_ids_by_repository_pull_request.sql
	queryListRunIDsByRepositoryPullRequest string
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

// ListRunIDsByRepositoryIssue returns run ids for one repository/issue pair.
func (r *Repository) ListRunIDsByRepositoryIssue(ctx context.Context, repositoryFullName string, issueNumber int64, limit int) ([]string, error) {
	return r.listRunIDsByRepositoryReference(ctx, queryListRunIDsByRepositoryIssue, repositoryFullName, issueNumber, limit, "issue_number", "repository/issue")
}

// ListRunIDsByRepositoryPullRequest returns run ids for one repository/pull request pair.
func (r *Repository) ListRunIDsByRepositoryPullRequest(ctx context.Context, repositoryFullName string, prNumber int64, limit int) ([]string, error) {
	return r.listRunIDsByRepositoryReference(ctx, queryListRunIDsByRepositoryPullRequest, repositoryFullName, prNumber, limit, "pr_number", "repository/pull request")
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
