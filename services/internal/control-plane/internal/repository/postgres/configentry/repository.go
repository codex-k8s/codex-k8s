package configentry

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/configentry"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

var (
	//go:embed sql/list.sql
	queryList string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/upsert_platform.sql
	queryUpsertPlatform string
	//go:embed sql/upsert_project.sql
	queryUpsertProject string
	//go:embed sql/upsert_repository.sql
	queryUpsertRepository string
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

	rows, err := r.db.Query(ctx, queryList, string(filter.Scope), filter.ProjectID, filter.RepositoryID, limit)
	if err != nil {
		return nil, fmt.Errorf("list config entries: %w", err)
	}
	defer rows.Close()

	out := make([]domainrepo.ConfigEntry, 0)
	for rows.Next() {
		var item domainrepo.ConfigEntry
		var scope string
		var kind string
		var mutability string
		if err := rows.Scan(
			&item.ID,
			&scope,
			&kind,
			&item.ProjectID,
			&item.RepositoryID,
			&item.Key,
			&item.Value,
			&item.SyncTargets,
			&mutability,
			&item.IsDangerous,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan config entry: %w", err)
		}
		item.Scope = enumtypes.ConfigEntryScope(strings.TrimSpace(scope))
		item.Kind = enumtypes.ConfigEntryKind(strings.TrimSpace(kind))
		item.Mutability = enumtypes.ConfigEntryMutability(strings.TrimSpace(mutability))
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate config entries: %w", err)
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (domainrepo.ConfigEntry, bool, error) {
	var item domainrepo.ConfigEntry
	var scope string
	var kind string
	var mutability string
	err := r.db.QueryRow(ctx, queryGetByID, id).Scan(
		&item.ID,
		&scope,
		&kind,
		&item.ProjectID,
		&item.RepositoryID,
		&item.Key,
		&item.Value,
		&item.SyncTargets,
		&mutability,
		&item.IsDangerous,
		&item.UpdatedAt,
	)
	if err == nil {
		item.Scope = enumtypes.ConfigEntryScope(strings.TrimSpace(scope))
		item.Kind = enumtypes.ConfigEntryKind(strings.TrimSpace(kind))
		item.Mutability = enumtypes.ConfigEntryMutability(strings.TrimSpace(mutability))
		return item, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.ConfigEntry{}, false, nil
	}
	return domainrepo.ConfigEntry{}, false, fmt.Errorf("get config entry by id: %w", err)
}

func (r *Repository) Exists(ctx context.Context, scope enumtypes.ConfigEntryScope, projectID string, repositoryID string, key string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, queryExists, string(scope), projectID, repositoryID, key).Scan(&exists); err != nil {
		return false, fmt.Errorf("check config entry exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.ConfigEntry, error) {
	var item domainrepo.ConfigEntry
	scope := enumtypes.ConfigEntryScope(strings.TrimSpace(string(params.Scope)))
	switch scope {
	case enumtypes.ConfigEntryScopePlatform:
		var itemScope string
		var itemKind string
		var itemMutability string
		err := r.db.QueryRow(
			ctx,
			queryUpsertPlatform,
			string(params.Kind),
			params.Key,
			params.ValuePlain,
			params.ValueEncrypted,
			params.SyncTargets,
			string(params.Mutability),
			params.IsDangerous,
			params.CreatedByUserID,
			params.UpdatedByUserID,
		).Scan(
			&item.ID,
			&itemScope,
			&itemKind,
			&item.ProjectID,
			&item.RepositoryID,
			&item.Key,
			&item.Value,
			&item.SyncTargets,
			&itemMutability,
			&item.IsDangerous,
			&item.UpdatedAt,
		)
		if err != nil {
			return domainrepo.ConfigEntry{}, fmt.Errorf("upsert config entry: %w", err)
		}
		item.Scope = enumtypes.ConfigEntryScope(strings.TrimSpace(itemScope))
		item.Kind = enumtypes.ConfigEntryKind(strings.TrimSpace(itemKind))
		item.Mutability = enumtypes.ConfigEntryMutability(strings.TrimSpace(itemMutability))
		return item, nil
	case enumtypes.ConfigEntryScopeProject:
		var itemScope string
		var itemKind string
		var itemMutability string
		err := r.db.QueryRow(
			ctx,
			queryUpsertProject,
			string(params.Kind),
			nullUUID(params.ProjectID),
			params.Key,
			params.ValuePlain,
			params.ValueEncrypted,
			params.SyncTargets,
			string(params.Mutability),
			params.IsDangerous,
			params.CreatedByUserID,
			params.UpdatedByUserID,
		).Scan(
			&item.ID,
			&itemScope,
			&itemKind,
			&item.ProjectID,
			&item.RepositoryID,
			&item.Key,
			&item.Value,
			&item.SyncTargets,
			&itemMutability,
			&item.IsDangerous,
			&item.UpdatedAt,
		)
		if err != nil {
			return domainrepo.ConfigEntry{}, fmt.Errorf("upsert config entry: %w", err)
		}
		item.Scope = enumtypes.ConfigEntryScope(strings.TrimSpace(itemScope))
		item.Kind = enumtypes.ConfigEntryKind(strings.TrimSpace(itemKind))
		item.Mutability = enumtypes.ConfigEntryMutability(strings.TrimSpace(itemMutability))
		return item, nil
	case enumtypes.ConfigEntryScopeRepository:
		var itemScope string
		var itemKind string
		var itemMutability string
		err := r.db.QueryRow(
			ctx,
			queryUpsertRepository,
			string(params.Kind),
			nullUUID(params.RepositoryID),
			params.Key,
			params.ValuePlain,
			params.ValueEncrypted,
			params.SyncTargets,
			string(params.Mutability),
			params.IsDangerous,
			params.CreatedByUserID,
			params.UpdatedByUserID,
		).Scan(
			&item.ID,
			&itemScope,
			&itemKind,
			&item.ProjectID,
			&item.RepositoryID,
			&item.Key,
			&item.Value,
			&item.SyncTargets,
			&itemMutability,
			&item.IsDangerous,
			&item.UpdatedAt,
		)
		if err != nil {
			return domainrepo.ConfigEntry{}, fmt.Errorf("upsert config entry: %w", err)
		}
		item.Scope = enumtypes.ConfigEntryScope(strings.TrimSpace(itemScope))
		item.Kind = enumtypes.ConfigEntryKind(strings.TrimSpace(itemKind))
		item.Mutability = enumtypes.ConfigEntryMutability(strings.TrimSpace(itemMutability))
		return item, nil
	default:
		return domainrepo.ConfigEntry{}, fmt.Errorf("upsert config entry: unsupported scope %q", string(params.Scope))
	}
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
