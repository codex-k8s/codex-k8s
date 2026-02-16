package configentry

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/configentry"
)

var (
	//go:embed sql/list.sql
	queryList string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/delete.sql
	queryDelete string
	//go:embed sql/exists.sql
	queryExists string
)

// Repository stores config entries in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filter domainrepo.ListFilter) ([]domainrepo.ConfigEntry, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}

	rows, err := r.db.Query(ctx, queryList, filter.Scope, filter.ProjectID, filter.RepositoryID, limit)
	if err != nil {
		return nil, fmt.Errorf("list config entries: %w", err)
	}
	defer rows.Close()

	out := make([]domainrepo.ConfigEntry, 0)
	for rows.Next() {
		var item domainrepo.ConfigEntry
		if err := rows.Scan(
			&item.ID,
			&item.Scope,
			&item.Kind,
			&item.ProjectID,
			&item.RepositoryID,
			&item.Key,
			&item.Value,
			&item.SyncTargets,
			&item.Mutability,
			&item.IsDangerous,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan config entry: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate config entries: %w", err)
	}
	return out, nil
}

func (r *Repository) Exists(ctx context.Context, scope string, projectID string, repositoryID string, key string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, queryExists, scope, projectID, repositoryID, key).Scan(&exists); err != nil {
		return false, fmt.Errorf("check config entry exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.ConfigEntry, error) {
	var item domainrepo.ConfigEntry
	err := r.db.QueryRow(
		ctx,
		queryUpsert,
		params.Scope,
		params.Kind,
		nullUUID(params.ProjectID),
		nullUUID(params.RepositoryID),
		params.Key,
		params.ValuePlain,
		params.ValueEncrypted,
		params.SyncTargets,
		params.Mutability,
		params.IsDangerous,
		nullUUID(params.CreatedByUserID),
		nullUUID(params.UpdatedByUserID),
	).Scan(
		&item.ID,
		&item.Scope,
		&item.Kind,
		&item.ProjectID,
		&item.RepositoryID,
		&item.Key,
		&item.Value,
		&item.SyncTargets,
		&item.Mutability,
		&item.IsDangerous,
		&item.UpdatedAt,
	)
	if err != nil {
		return domainrepo.ConfigEntry{}, fmt.Errorf("upsert config entry: %w", err)
	}
	return item, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return postgres.ExecRequireRowOrWrap(ctx, r.db, queryDelete, "delete config entry", id)
}

func nullUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}
