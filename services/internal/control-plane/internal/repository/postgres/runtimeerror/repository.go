package runtimeerror

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimeerror"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/runtimeerror/dbmodel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/insert.sql
	queryInsert string
	//go:embed sql/list_all.sql
	queryListAll string
	//go:embed sql/list_for_user.sql
	queryListForUser string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/mark_viewed.sql
	queryMarkViewed string
)

// Repository persists runtime_errors in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs runtime errors repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Insert appends one runtime error entry.
func (r *Repository) Insert(ctx context.Context, params domainrepo.RecordParams) (domainrepo.Item, error) {
	rows, err := r.db.Query(
		ctx,
		queryInsert,
		strings.TrimSpace(params.Source),
		strings.TrimSpace(params.Level),
		strings.TrimSpace(params.Message),
		jsonOrEmptyObject(params.DetailsJSON),
		nullableText(params.StackTrace),
		nullableText(params.CorrelationID),
		nullableUUID(params.RunID),
		nullableUUID(params.ProjectID),
		nullableText(params.Namespace),
		nullableText(params.JobName),
	)
	if err != nil {
		return domainrepo.Item{}, fmt.Errorf("insert runtime error: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.RuntimeErrorRow])
	if err != nil {
		return domainrepo.Item{}, fmt.Errorf("collect inserted runtime error: %w", err)
	}
	return fromDBModel(row), nil
}

// ListAll returns runtime errors for platform admin scope.
func (r *Repository) ListAll(ctx context.Context, filter domainrepo.ListFilter) ([]domainrepo.Item, error) {
	f := normalizeListFilter(filter)
	rows, err := r.db.Query(ctx, queryListAll, f.Limit, string(f.State), nullableText(f.Level), nullableText(f.Source), nullableText(f.RunID), nullableText(f.CorrelationID))
	if err != nil {
		return nil, fmt.Errorf("list runtime errors: %w", err)
	}
	return collectItems(rows, "runtime errors")
}

// ListForUser returns runtime errors visible to user projects.
func (r *Repository) ListForUser(ctx context.Context, userID string, filter domainrepo.ListFilter) ([]domainrepo.Item, error) {
	f := normalizeListFilter(filter)
	rows, err := r.db.Query(
		ctx,
		queryListForUser,
		strings.TrimSpace(userID),
		f.Limit,
		string(f.State),
		nullableText(f.Level),
		nullableText(f.Source),
		nullableText(f.RunID),
		nullableText(f.CorrelationID),
	)
	if err != nil {
		return nil, fmt.Errorf("list runtime errors for user: %w", err)
	}
	return collectItems(rows, "runtime errors for user")
}

// GetByID returns one runtime error by id.
func (r *Repository) GetByID(ctx context.Context, id string) (domainrepo.Item, bool, error) {
	rows, err := r.db.Query(ctx, queryGetByID, strings.TrimSpace(id))
	if err != nil {
		return domainrepo.Item{}, false, fmt.Errorf("query runtime error by id: %w", err)
	}
	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.RuntimeErrorRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Item{}, false, nil
		}
		return domainrepo.Item{}, false, fmt.Errorf("collect runtime error by id: %w", err)
	}
	return fromDBModel(item), true, nil
}

// MarkViewed marks one runtime error as viewed.
func (r *Repository) MarkViewed(ctx context.Context, params domainrepo.MarkViewedParams) (domainrepo.Item, bool, error) {
	rows, err := r.db.Query(ctx, queryMarkViewed, strings.TrimSpace(params.ID), nullableUUID(params.ViewerID))
	if err != nil {
		return domainrepo.Item{}, false, fmt.Errorf("mark runtime error viewed: %w", err)
	}
	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.RuntimeErrorRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Item{}, false, nil
		}
		return domainrepo.Item{}, false, fmt.Errorf("collect viewed runtime error: %w", err)
	}
	return fromDBModel(item), true, nil
}

func collectItems(rows pgx.Rows, op string) ([]domainrepo.Item, error) {
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.RuntimeErrorRow])
	if err != nil {
		return nil, fmt.Errorf("collect %s: %w", op, err)
	}
	out := make([]domainrepo.Item, 0, len(items))
	for _, row := range items {
		out = append(out, fromDBModel(row))
	}
	return out, nil
}

func normalizeListFilter(filter domainrepo.ListFilter) domainrepo.ListFilter {
	normalized := filter
	if normalized.Limit <= 0 {
		normalized.Limit = 100
	}
	if normalized.Limit > 1000 {
		normalized.Limit = 1000
	}
	normalized.Level = strings.TrimSpace(normalized.Level)
	normalized.Source = strings.TrimSpace(normalized.Source)
	normalized.RunID = strings.TrimSpace(normalized.RunID)
	normalized.CorrelationID = strings.TrimSpace(normalized.CorrelationID)
	normalized.State = normalizeListState(normalized.State)
	return normalized
}

func normalizeListState(state querytypes.RuntimeErrorListState) querytypes.RuntimeErrorListState {
	switch state {
	case querytypes.RuntimeErrorListStateActive, querytypes.RuntimeErrorListStateViewed, querytypes.RuntimeErrorListStateAll:
		return state
	default:
		return querytypes.RuntimeErrorListStateActive
	}
}

func nullableText(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableUUID(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func jsonOrEmptyObject(value []byte) []byte {
	if len(value) == 0 {
		return []byte(`{}`)
	}
	return value
}
