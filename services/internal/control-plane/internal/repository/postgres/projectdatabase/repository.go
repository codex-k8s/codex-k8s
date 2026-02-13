package projectdatabase

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectdatabase"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/repository/postgres/projectdatabase/dbmodel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/get_by_database_name.sql
	queryGetByDatabaseName string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/delete_by_database_name.sql
	queryDeleteByDatabaseName string
)

// Repository stores project_databases rows in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL project database ownership repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetByDatabaseName returns one ownership row by database name.
func (r *Repository) GetByDatabaseName(ctx context.Context, databaseName string) (domainrepo.Item, bool, error) {
	var row dbmodel.ProjectDatabaseRow
	err := r.db.QueryRow(ctx, queryGetByDatabaseName, databaseName).Scan(
		&row.ProjectID,
		&row.Environment,
		&row.DatabaseName,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err == nil {
		return fromDBModel(row), true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.Item{}, false, nil
	}
	return domainrepo.Item{}, false, fmt.Errorf("query project database by name: %w", err)
}

// Upsert creates or updates ownership mapping.
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.Item, error) {
	rows, err := r.db.Query(ctx, queryUpsert, params.ProjectID, params.Environment, params.DatabaseName)
	if err != nil {
		return domainrepo.Item{}, fmt.Errorf("upsert project database ownership: %w", err)
	}
	item, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dbmodel.ProjectDatabaseRow])
	if err != nil {
		return domainrepo.Item{}, fmt.Errorf("collect upserted project database ownership: %w", err)
	}
	return fromDBModel(item), nil
}

// DeleteByDatabaseName removes ownership mapping by database name.
func (r *Repository) DeleteByDatabaseName(ctx context.Context, databaseName string) (bool, error) {
	var deleted bool
	if err := r.db.QueryRow(ctx, queryDeleteByDatabaseName, databaseName).Scan(&deleted); err != nil {
		return false, fmt.Errorf("delete project database ownership: %w", err)
	}
	return deleted, nil
}
