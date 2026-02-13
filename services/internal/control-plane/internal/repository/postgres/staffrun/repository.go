package staffrun

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

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
	//go:embed sql/get_by_id.sql
	queryGetByID string
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

// GetByID returns a run by id.
func (r *Repository) GetByID(ctx context.Context, runID string) (domainrepo.Run, bool, error) {
	rows, err := r.db.Query(ctx, queryGetByID, runID)
	if err != nil {
		return domainrepo.Run{}, false, fmt.Errorf("query run by id: %w", err)
	}
	runRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RunRow])
	if err != nil {
		return domainrepo.Run{}, false, fmt.Errorf("collect run by id: %w", err)
	}
	if len(runRows) == 0 {
		return domainrepo.Run{}, false, nil
	}
	return runFromDBModel(runRows[0]), true, nil
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
