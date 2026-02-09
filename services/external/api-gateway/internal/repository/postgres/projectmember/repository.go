package projectmember

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/projectmember"
)

var (
	//go:embed sql/list.sql
	queryList string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/get_role.sql
	queryGetRole string
)

// Repository stores project members in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL project member repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns members for a project.
func (r *Repository) List(ctx context.Context, projectID string, limit int) ([]domainrepo.Member, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryList, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}
	defer rows.Close()

	var out []domainrepo.Member
	for rows.Next() {
		var m domainrepo.Member
		if err := rows.Scan(&m.ProjectID, &m.UserID, &m.Email, &m.Role); err != nil {
			return nil, fmt.Errorf("scan project member: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project members: %w", err)
	}
	return out, nil
}

// Upsert sets membership role for a user.
func (r *Repository) Upsert(ctx context.Context, projectID string, userID string, role string) error {
	if _, err := r.db.ExecContext(ctx, queryUpsert, projectID, userID, role); err != nil {
		return fmt.Errorf("upsert project member: %w", err)
	}
	return nil
}

// GetRole returns role for a project member.
func (r *Repository) GetRole(ctx context.Context, projectID string, userID string) (string, bool, error) {
	var role string
	err := r.db.QueryRowContext(ctx, queryGetRole, projectID, userID).Scan(&role)
	if err == nil {
		return role, true, nil
	}
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return "", false, fmt.Errorf("get project member role: %w", err)
}
