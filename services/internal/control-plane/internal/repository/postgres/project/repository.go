package project

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
)

var (
	//go:embed sql/list_all.sql
	queryListAll string
	//go:embed sql/list_for_user.sql
	queryListForUser string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/get_by_id.sql
	queryGetByID string
	//go:embed sql/delete_by_id.sql
	queryDeleteByID string
	//go:embed sql/get_learning_mode_default.sql
	queryGetLearningModeDefault string
)

// Repository stores projects in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL project repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ListAll returns all projects.
func (r *Repository) ListAll(ctx context.Context, limit int) ([]domainrepo.Project, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, queryListAll, limit)
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
	rows, err := r.db.Query(ctx, queryListForUser, userID, limit)
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

// Upsert creates/updates a project by slug.
func (r *Repository) Upsert(ctx context.Context, params domainrepo.UpsertParams) (domainrepo.Project, error) {
	var out domainrepo.Project
	err := r.db.QueryRow(ctx, queryUpsert, params.ID, params.Slug, params.Name, params.SettingsJSON).Scan(&out.ID, &out.Slug, &out.Name)
	if err != nil {
		return domainrepo.Project{}, fmt.Errorf("upsert project: %w", err)
	}
	return out, nil
}

// GetByID returns a project by id.
func (r *Repository) GetByID(ctx context.Context, projectID string) (domainrepo.Project, bool, error) {
	var p domainrepo.Project
	err := r.db.QueryRow(ctx, queryGetByID, projectID).Scan(&p.ID, &p.Slug, &p.Name)
	if err == nil {
		return p, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.Project{}, false, nil
	}
	return domainrepo.Project{}, false, fmt.Errorf("get project by id: %w", err)
}

// DeleteByID deletes a project by id.
func (r *Repository) DeleteByID(ctx context.Context, projectID string) error {
	return postgres.ExecRequireRowOrWrap(ctx, r.db, queryDeleteByID, "delete project by id", projectID)
}

// GetLearningModeDefault returns project default learning-mode flag from JSONB settings.
func (r *Repository) GetLearningModeDefault(ctx context.Context, projectID string) (bool, bool, error) {
	var enabled bool
	err := r.db.QueryRow(ctx, queryGetLearningModeDefault, projectID).Scan(&enabled)
	if err == nil {
		return enabled, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	return false, false, fmt.Errorf("get project learning_mode_default: %w", err)
}
