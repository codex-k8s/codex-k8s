package staffrun

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/staffrun"
)

var (
	//go:embed sql/list_all.sql
	queryListAll string
	//go:embed sql/list_for_user.sql
	queryListForUser string
	//go:embed sql/list_events_by_correlation.sql
	queryListEventsByCorrelation string
	//go:embed sql/get_correlation_by_run_id.sql
	queryGetCorrelationByRunID string
)

// Repository loads runs and flow events from PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs staff run repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListAll returns recent runs for platform admins.
func (r *Repository) ListAll(ctx context.Context, limit int) ([]domainrepo.Run, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryListAll, limit)
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}
	defer rows.Close()

	var out []domainrepo.Run
	for rows.Next() {
		var item domainrepo.Run
		if err := rows.Scan(&item.ID, &item.CorrelationID, &item.ProjectID, &item.Status, &item.CreatedAt, &item.StartedAt, &item.FinishedAt); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runs: %w", err)
	}
	return out, nil
}

// ListForUser returns runs for projects the user is a member of.
func (r *Repository) ListForUser(ctx context.Context, userID string, limit int) ([]domainrepo.Run, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryListForUser, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list runs for user: %w", err)
	}
	defer rows.Close()

	var out []domainrepo.Run
	for rows.Next() {
		var item domainrepo.Run
		if err := rows.Scan(&item.ID, &item.CorrelationID, &item.ProjectID, &item.Status, &item.CreatedAt, &item.StartedAt, &item.FinishedAt); err != nil {
			return nil, fmt.Errorf("scan run for user: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runs for user: %w", err)
	}
	return out, nil
}

// ListEventsByCorrelation returns events for a correlation id.
func (r *Repository) ListEventsByCorrelation(ctx context.Context, correlationID string, limit int) ([]domainrepo.FlowEvent, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryListEventsByCorrelation, correlationID, limit)
	if err != nil {
		return nil, fmt.Errorf("list flow events: %w", err)
	}
	defer rows.Close()

	var out []domainrepo.FlowEvent
	for rows.Next() {
		var payloadText string
		var item domainrepo.FlowEvent
		if err := rows.Scan(&item.CorrelationID, &item.EventType, &item.CreatedAt, &payloadText); err != nil {
			return nil, fmt.Errorf("scan flow event: %w", err)
		}
		item.PayloadJSON = []byte(payloadText)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate flow events: %w", err)
	}
	return out, nil
}

// GetCorrelationByRunID returns correlation id and project id for a run id.
func (r *Repository) GetCorrelationByRunID(ctx context.Context, runID string) (string, string, bool, error) {
	var correlationID string
	var projectID string
	err := r.db.QueryRowContext(ctx, queryGetCorrelationByRunID, runID).Scan(&correlationID, &projectID)
	if err == nil {
		return correlationID, projectID, true, nil
	}
	if err == sql.ErrNoRows {
		return "", "", false, nil
	}
	return "", "", false, fmt.Errorf("get correlation by run id: %w", err)
}
