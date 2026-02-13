package mcpactionrequest

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/mcpactionrequest"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/mcpactionrequest/dbmodel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/create.sql
	queryCreate string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/find_latest_by_signature.sql
	queryFindLatestBySignature string
	//go:embed sql/find_pending_by_signature.sql
	queryFindPendingBySignature string
	//go:embed sql/list_pending.sql
	queryListPending string
	//go:embed sql/update_state.sql
	queryUpdateState string
)

// Repository persists mcp_action_requests in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL mcp_action_requests repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts a new action request row.
func (r *Repository) Create(ctx context.Context, params domainrepo.CreateParams) (domainrepo.Item, error) {
	targetRef := jsonOrEmptyObject(params.TargetRef)
	payload := jsonOrEmptyObject(params.Payload)

	var id int64
	if err := r.db.QueryRow(
		ctx,
		queryCreate,
		params.CorrelationID,
		nullableUUID(params.RunID),
		params.ToolName,
		params.Action,
		targetRef,
		string(params.ApprovalMode),
		string(params.ApprovalState),
		params.RequestedBy,
		nullableText(params.AppliedBy),
		payload,
	).Scan(&id); err != nil {
		return domainrepo.Item{}, fmt.Errorf("create mcp action request: %w", err)
	}

	item, ok, err := r.GetByID(ctx, id)
	if err != nil {
		return domainrepo.Item{}, err
	}
	if !ok {
		return domainrepo.Item{}, fmt.Errorf("created mcp action request %d not found", id)
	}
	return item, nil
}

// GetByID returns one request by id.
func (r *Repository) GetByID(ctx context.Context, id int64) (domainrepo.Item, bool, error) {
	rows, err := r.db.Query(ctx, queryGetByID, id)
	if err != nil {
		return domainrepo.Item{}, false, fmt.Errorf("query mcp action request by id: %w", err)
	}
	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.ActionRequestRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Item{}, false, nil
		}
		return domainrepo.Item{}, false, fmt.Errorf("collect mcp action request by id: %w", err)
	}
	if item.ID == 0 {
		return domainrepo.Item{}, false, nil
	}
	return fromDBModel(item), true, nil
}

// FindLatestBySignature returns latest request for idempotent retries.
func (r *Repository) FindLatestBySignature(ctx context.Context, runID string, toolName string, action string, targetRefJSON []byte) (domainrepo.Item, bool, error) {
	return r.findBySignature(ctx, runID, toolName, action, targetRefJSON, queryFindLatestBySignature, "latest")
}

// FindPendingBySignature returns latest pending request for idempotent retries.
func (r *Repository) FindPendingBySignature(ctx context.Context, runID string, toolName string, action string, targetRefJSON []byte) (domainrepo.Item, bool, error) {
	return r.findBySignature(ctx, runID, toolName, action, targetRefJSON, queryFindPendingBySignature, "pending")
}

func (r *Repository) findBySignature(ctx context.Context, runID string, toolName string, action string, targetRefJSON []byte, sqlQuery string, queryKind string) (domainrepo.Item, bool, error) {
	if runID == "" {
		return domainrepo.Item{}, false, nil
	}
	rows, err := r.db.Query(ctx, sqlQuery, runID, toolName, action, jsonOrEmptyObject(targetRefJSON))
	if err != nil {
		return domainrepo.Item{}, false, fmt.Errorf("query %s mcp action request by signature: %w", queryKind, err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.ActionRequestRow])
	if err != nil {
		return domainrepo.Item{}, false, fmt.Errorf("collect %s mcp action request by signature: %w", queryKind, err)
	}
	if len(items) == 0 {
		return domainrepo.Item{}, false, nil
	}
	return fromDBModel(items[0]), true, nil
}

// ListPending returns pending approval queue.
func (r *Repository) ListPending(ctx context.Context, limit int) ([]domainrepo.Item, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.Query(ctx, queryListPending, limit)
	if err != nil {
		return nil, fmt.Errorf("query pending mcp action requests: %w", err)
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[dbmodel.ActionRequestRow])
	if err != nil {
		return nil, fmt.Errorf("collect pending mcp action requests: %w", err)
	}
	out := make([]domainrepo.Item, 0, len(items))
	for _, item := range items {
		out = append(out, fromDBModel(item))
	}
	return out, nil
}

// UpdateState updates approval state and returns updated row.
func (r *Repository) UpdateState(ctx context.Context, params domainrepo.UpdateStateParams) (domainrepo.Item, bool, error) {
	var id int64
	err := r.db.QueryRow(
		ctx,
		queryUpdateState,
		params.ID,
		string(params.ApprovalState),
		nullableText(params.AppliedBy),
		jsonOrEmptyObject(params.Payload),
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Item{}, false, nil
		}
		return domainrepo.Item{}, false, fmt.Errorf("update mcp action request state: %w", err)
	}

	item, ok, err := r.GetByID(ctx, id)
	if err != nil {
		return domainrepo.Item{}, false, err
	}
	if !ok {
		return domainrepo.Item{}, false, nil
	}
	return item, true, nil
}

func nullableUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func jsonOrEmptyObject(raw []byte) []byte {
	if len(raw) == 0 {
		return []byte(`{}`)
	}
	return raw
}
