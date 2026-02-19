package runaccesskey

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runaccesskey"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/runaccesskey/dbmodel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/get_by_run_id.sql
	queryGetByRunID string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/revoke.sql
	queryRevoke string
	//go:embed sql/touch_last_used.sql
	queryTouchLastUsed string
)

// Repository persists run access keys in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs run access key repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetByRunID returns one run access key by run id.
func (r *Repository) GetByRunID(ctx context.Context, runID string) (domainrepo.Run, bool, error) {
	rows, err := r.db.Query(ctx, queryGetByRunID, strings.TrimSpace(runID))
	if err != nil {
		return domainrepo.Run{}, false, fmt.Errorf("query run access key by run id: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.RunAccessKeyRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Run{}, false, nil
		}
		return domainrepo.Run{}, false, fmt.Errorf("collect run access key by run id: %w", err)
	}
	return fromDBModel(row), true, nil
}

// Upsert creates or updates run access key row.
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.Run, error) {
	now := params.UpdatedAt.UTC()
	rows, err := r.db.Query(
		ctx,
		queryUpsert,
		strings.TrimSpace(params.RunID),
		nullableUUID(params.ProjectID),
		strings.TrimSpace(params.CorrelationID),
		strings.TrimSpace(params.RuntimeMode),
		nullableText(params.Namespace),
		nullableText(params.TargetEnv),
		params.KeyHash,
		normalizeStatus(params.Status),
		params.IssuedAt.UTC(),
		params.ExpiresAt.UTC(),
		nullableTime(params.RevokedAt),
		nullableTime(params.LastUsedAt),
		strings.TrimSpace(params.CreatedBy),
		now,
		now,
	)
	if err != nil {
		return domainrepo.Run{}, fmt.Errorf("upsert run access key: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.RunAccessKeyRow])
	if err != nil {
		return domainrepo.Run{}, fmt.Errorf("collect upserted run access key: %w", err)
	}
	return fromDBModel(row), nil
}

// Revoke marks run access key as revoked.
func (r *Repository) Revoke(ctx context.Context, runID string, revokedAt time.Time, updatedAt time.Time) (domainrepo.Run, bool, error) {
	rows, err := r.db.Query(ctx, queryRevoke, strings.TrimSpace(runID), revokedAt.UTC(), updatedAt.UTC())
	if err != nil {
		return domainrepo.Run{}, false, fmt.Errorf("revoke run access key: %w", err)
	}
	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.RunAccessKeyRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrepo.Run{}, false, nil
		}
		return domainrepo.Run{}, false, fmt.Errorf("collect revoked run access key: %w", err)
	}
	return fromDBModel(row), true, nil
}

// TouchLastUsed updates last_used_at for active key.
func (r *Repository) TouchLastUsed(ctx context.Context, runID string, usedAt time.Time) error {
	if _, err := r.db.Exec(ctx, queryTouchLastUsed, strings.TrimSpace(runID), usedAt.UTC(), usedAt.UTC()); err != nil {
		return fmt.Errorf("touch run access key last_used_at: %w", err)
	}
	return nil
}

func fromDBModel(row dbmodel.RunAccessKeyRow) domainrepo.Run {
	return domainrepo.Run{
		RunID:         strings.TrimSpace(row.RunID),
		ProjectID:     textOrEmpty(row.ProjectID),
		CorrelationID: strings.TrimSpace(row.CorrelationID),
		RuntimeMode:   strings.TrimSpace(row.RuntimeMode),
		Namespace:     textOrEmpty(row.Namespace),
		TargetEnv:     textOrEmpty(row.TargetEnv),
		KeyHash:       append([]byte(nil), row.KeyHash...),
		Status:        normalizeStatus(strings.TrimSpace(row.Status)),
		IssuedAt:      row.IssuedAt.UTC(),
		ExpiresAt:     row.ExpiresAt.UTC(),
		RevokedAt:     timestamptzToTimePtr(row.RevokedAt),
		LastUsedAt:    timestamptzToTimePtr(row.LastUsedAt),
		CreatedBy:     strings.TrimSpace(row.CreatedBy),
		CreatedAt:     row.CreatedAt.UTC(),
		UpdatedAt:     row.UpdatedAt.UTC(),
	}
}

func normalizeStatus(raw string) entitytypes.RunAccessKeyStatus {
	value := entitytypes.RunAccessKeyStatus(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case entitytypes.RunAccessKeyStatusRevoked:
		return entitytypes.RunAccessKeyStatusRevoked
	default:
		return entitytypes.RunAccessKeyStatusActive
	}
}

func nullableUUID(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableText(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableTime(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	result := value.UTC()
	return result
}

func textOrEmpty(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}

func timestamptzToTimePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	out := value.Time.UTC()
	return &out
}
