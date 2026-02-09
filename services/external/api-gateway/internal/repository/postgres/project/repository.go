package project

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/project"
)

var (
	//go:embed sql/list_all.sql
	queryListAll string
	//go:embed sql/list_for_user.sql
	queryListForUser string
)

// Repository stores projects in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL project repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListAll returns all projects.
func (r *Repository) ListAll(ctx context.Context, limit int) ([]domainrepo.Project, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryListAll, limit)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var out []domainrepo.Project
	for rows.Next() {
		var p domainrepo.Project
		if err := rows.Scan(&p.ID, &p.Slug, &p.Name); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return out, nil
}

// ListForUser returns projects visible for a user.
func (r *Repository) ListForUser(ctx context.Context, userID string, limit int) ([]domainrepo.ProjectWithRole, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryListForUser, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list projects for user: %w", err)
	}
	defer rows.Close()

	var out []domainrepo.ProjectWithRole
	for rows.Next() {
		var p domainrepo.ProjectWithRole
		if err := rows.Scan(&p.ID, &p.Slug, &p.Name, &p.Role); err != nil {
			return nil, fmt.Errorf("scan project for user: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects for user: %w", err)
	}
	return out, nil
}

