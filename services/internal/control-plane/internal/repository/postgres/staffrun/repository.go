package staffrun

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/staffrun/dbmodel"
)

var (
	//go:embed sql/list_all.sql
	queryListAll string
	//go:embed sql/list_for_user.sql
	queryListForUser string
	//go:embed sql/list_jobs_all.sql
	queryListJobsAll string
	//go:embed sql/list_jobs_for_user.sql
	queryListJobsForUser string
	//go:embed sql/list_waits_all.sql
	queryListWaitsAll string
	//go:embed sql/list_waits_for_user.sql
	queryListWaitsForUser string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/get_logs_by_run_id.sql
	queryGetLogsByRunID string
	//go:embed sql/list_events_by_correlation.sql
	queryListEventsByCorrelation string
	//go:embed sql/delete_events_by_project_id.sql
	queryDeleteEventsByProjectID string
	//go:embed sql/get_correlation_by_run_id.sql
	queryGetCorrelationByRunID string
)

// Repository loads runs and flow events from PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs staff run repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ListAll returns recent runs for platform admins.
func (r *Repository) ListAll(ctx context.Context, limit int) ([]domainrepo.Run, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, queryListAll, limit)
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}
	return collectRuns(rows, "runs")
}

// ListForUser returns runs for projects the user is a member of.
func (r *Repository) ListForUser(ctx context.Context, userID string, limit int) ([]domainrepo.Run, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, queryListForUser, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list runs for user: %w", err)
	}
	return collectRuns(rows, "runs for user")
}

// ListJobsAll returns runtime jobs list for platform admins.
func (r *Repository) ListJobsAll(ctx context.Context, filter domainrepo.ListFilter) ([]domainrepo.Run, error) {
	args := normalizeListFilter(filter)
	rows, err := r.db.Query(ctx, queryListJobsAll, args.Limit, args.TriggerKind, args.Status, args.AgentKey)
	if err != nil {
		return nil, fmt.Errorf("list run jobs: %w", err)
	}
	return collectRuns(rows, "run jobs")
}

// ListJobsForUser returns runtime jobs list scoped to user projects.
func (r *Repository) ListJobsForUser(ctx context.Context, userID string, filter domainrepo.ListFilter) ([]domainrepo.Run, error) {
	args := normalizeListFilter(filter)
	rows, err := r.db.Query(ctx, queryListJobsForUser, userID, args.Limit, args.TriggerKind, args.Status, args.AgentKey)
	if err != nil {
		return nil, fmt.Errorf("list run jobs for user: %w", err)
	}
	return collectRuns(rows, "run jobs for user")
}

// ListWaitsAll returns wait queue list for platform admins.
func (r *Repository) ListWaitsAll(ctx context.Context, filter domainrepo.ListFilter) ([]domainrepo.Run, error) {
	args := normalizeListFilter(filter)
	rows, err := r.db.Query(ctx, queryListWaitsAll, args.Limit, args.TriggerKind, args.Status, args.AgentKey, args.WaitState)
	if err != nil {
		return nil, fmt.Errorf("list run waits: %w", err)
	}
	return collectRuns(rows, "run waits")
}

// ListWaitsForUser returns wait queue list scoped to user projects.
func (r *Repository) ListWaitsForUser(ctx context.Context, userID string, filter domainrepo.ListFilter) ([]domainrepo.Run, error) {
	args := normalizeListFilter(filter)
	rows, err := r.db.Query(ctx, queryListWaitsForUser, userID, args.Limit, args.TriggerKind, args.Status, args.AgentKey, args.WaitState)
	if err != nil {
		return nil, fmt.Errorf("list run waits for user: %w", err)
	}
	return collectRuns(rows, "run waits for user")
}

// GetByID returns a run by id.
func (r *Repository) GetByID(ctx context.Context, runID string) (domainrepo.Run, bool, error) {
	return queryOneMappedByID(ctx, r.db, queryGetByID, runID, "run", runFromDBModel)
}

// GetLogsByRunID returns one run logs snapshot by run id.
func (r *Repository) GetLogsByRunID(ctx context.Context, runID string) (domainrepo.RunLogs, bool, error) {
	return queryOneMappedByID(ctx, r.db, queryGetLogsByRunID, runID, "run logs", runLogsFromDBModel)
}

// ListEventsByCorrelation returns events for a correlation id.
func (r *Repository) ListEventsByCorrelation(ctx context.Context, correlationID string, limit int) ([]domainrepo.FlowEvent, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, queryListEventsByCorrelation, correlationID, limit)
	if err != nil {
		return nil, fmt.Errorf("list flow events: %w", err)
	}

	eventRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.FlowEventRow])
	if err != nil {
		return nil, fmt.Errorf("collect flow events: %w", err)
	}
	out := make([]domainrepo.FlowEvent, 0, len(eventRows))
	for _, eventRow := range eventRows {
		out = append(out, flowEventFromDBModel(eventRow))
	}
	return out, nil
}

// DeleteFlowEventsByProjectID removes flow events for all runs of a project.
func (r *Repository) DeleteFlowEventsByProjectID(ctx context.Context, projectID string) error {
	if projectID == "" {
		return nil
	}
	if _, err := r.db.Exec(ctx, queryDeleteEventsByProjectID, projectID); err != nil {
		return fmt.Errorf("delete flow events by project id: %w", err)
	}
	return nil
}

// GetCorrelationByRunID returns correlation id and project id for a run id.
func (r *Repository) GetCorrelationByRunID(ctx context.Context, runID string) (string, string, bool, error) {
	var correlationID string
	var projectID string
	err := r.db.QueryRow(ctx, queryGetCorrelationByRunID, runID).Scan(&correlationID, &projectID)
	if err == nil {
		return correlationID, projectID, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, nil
	}
	return "", "", false, fmt.Errorf("get correlation by run id: %w", err)
}

func collectRuns(rows pgx.Rows, operationLabel string) ([]domainrepo.Run, error) {
	runRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RunRow])
	if err != nil {
		return nil, fmt.Errorf("collect %s: %w", operationLabel, err)
	}
	out := make([]domainrepo.Run, 0, len(runRows))
	for _, runRow := range runRows {
		out = append(out, runFromDBModel(runRow))
	}
	return out, nil
}

func normalizeListFilter(filter domainrepo.ListFilter) domainrepo.ListFilter {
	normalized := filter
	if normalized.Limit <= 0 {
		normalized.Limit = 200
	}
	if normalized.Limit > 1000 {
		normalized.Limit = 1000
	}
	normalized.TriggerKind = strings.TrimSpace(normalized.TriggerKind)
	normalized.Status = strings.TrimSpace(normalized.Status)
	normalized.AgentKey = strings.TrimSpace(normalized.AgentKey)
	normalized.WaitState = strings.TrimSpace(normalized.WaitState)
	return normalized
}

func queryOneRowByID[T any](ctx context.Context, db *pgxpool.Pool, query string, id string, operationLabel string) (T, bool, error) {
	rows, err := db.Query(ctx, query, id)
	if err != nil {
		var zero T
		return zero, false, fmt.Errorf("query %s by id: %w", operationLabel, err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		var zero T
		return zero, false, fmt.Errorf("collect %s by id: %w", operationLabel, err)
	}
	if len(items) == 0 {
		var zero T
		return zero, false, nil
	}
	return items[0], true, nil
}

func queryOneMappedByID[T any, Out any](
	ctx context.Context,
	db *pgxpool.Pool,
	query string,
	id string,
	operationLabel string,
	cast func(T) Out,
) (Out, bool, error) {
	row, ok, err := queryOneRowByID[T](ctx, db, query, id, operationLabel)
	if err != nil {
		var zero Out
		return zero, false, err
	}
	if !ok {
		var zero Out
		return zero, false, nil
	}
	return cast(row), true, nil
}
