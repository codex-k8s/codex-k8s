package realtime

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/get_event_by_id.sql
	queryGetEventByID string
	//go:embed sql/list_events_after_id.sql
	queryListEventsAfterID string
	//go:embed sql/check_project_membership.sql
	queryCheckProjectMembership string
	//go:embed sql/cleanup_old_events.sql
	queryCleanupOldEvents string
)

type rowScanner interface {
	Scan(dest ...any) error
}

// Repository reads realtime events and performs lightweight RBAC checks.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs realtime events repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Event, bool, error) {
	if r == nil || r.db == nil {
		return Event{}, false, fmt.Errorf("realtime repository is not configured")
	}
	if id <= 0 {
		return Event{}, false, nil
	}
	row := r.db.QueryRow(ctx, queryGetEventByID, id)
	item, err := scanEvent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Event{}, false, nil
		}
		return Event{}, false, fmt.Errorf("query realtime event by id=%d: %w", id, err)
	}
	return item, true, nil
}

func (r *Repository) ListAfterID(ctx context.Context, afterID int64, limit int) ([]Event, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("realtime repository is not configured")
	}
	if afterID < 0 {
		afterID = 0
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 2000 {
		limit = 2000
	}

	rows, err := r.db.Query(ctx, queryListEventsAfterID, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list realtime events after id=%d: %w", afterID, err)
	}
	defer rows.Close()

	items := make([]Event, 0, limit)
	for rows.Next() {
		item, scanErr := scanEvent(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan realtime event row: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate realtime events rows: %w", err)
	}
	return items, nil
}

func (r *Repository) UserHasProjectAccess(ctx context.Context, projectID string, userID string) (bool, error) {
	if r == nil || r.db == nil {
		return false, fmt.Errorf("realtime repository is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	userID = strings.TrimSpace(userID)
	if projectID == "" || userID == "" {
		return false, nil
	}
	var allowed bool
	if err := r.db.QueryRow(ctx, queryCheckProjectMembership, projectID, userID).Scan(&allowed); err != nil {
		return false, fmt.Errorf("check realtime project access project_id=%s user_id=%s: %w", projectID, userID, err)
	}
	return allowed, nil
}

func (r *Repository) CleanupOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("realtime repository is not configured")
	}
	tag, err := r.db.Exec(ctx, queryCleanupOldEvents, cutoff.UTC())
	if err != nil {
		return 0, fmt.Errorf("cleanup realtime events older than %s: %w", cutoff.UTC().Format(time.RFC3339), err)
	}
	return tag.RowsAffected(), nil
}

func scanEvent(row rowScanner) (Event, error) {
	var item Event
	if err := row.Scan(
		&item.ID,
		&item.Topic,
		&item.ScopeJSON,
		&item.PayloadJSON,
		&item.CorrelationID,
		&item.ProjectID,
		&item.RunID,
		&item.TaskID,
		&item.CreatedAt,
	); err != nil {
		return Event{}, err
	}
	item.Topic = strings.TrimSpace(item.Topic)
	item.CorrelationID = strings.TrimSpace(item.CorrelationID)
	item.ProjectID = strings.TrimSpace(item.ProjectID)
	item.RunID = strings.TrimSpace(item.RunID)
	item.TaskID = strings.TrimSpace(item.TaskID)
	return item, nil
}
