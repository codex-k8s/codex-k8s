package projectmember

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/codex-k8s/codex-k8s/libs/go/postgres"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
)

var (
	//go:embed sql/list.sql
	queryList string
	//go:embed sql/upsert.sql
	queryUpsert string
	//go:embed sql/delete.sql
	queryDelete string
	//go:embed sql/get_role.sql
	queryGetRole string
	//go:embed sql/set_learning_mode_override.sql
	querySetLearningModeOverride string
	//go:embed sql/get_learning_mode_override.sql
	queryGetLearningModeOverride string
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
	defer func() { _ = rows.Close() }()

	var out []domainrepo.Member
	for rows.Next() {
		var (
			m        domainrepo.Member
			override sql.NullBool
		)
		if err := rows.Scan(&m.ProjectID, &m.UserID, &m.Email, &m.Role, &override); err != nil {
			return nil, fmt.Errorf("scan project member: %w", err)
		}
		if override.Valid {
			v := override.Bool
			m.LearningModeOverride = &v
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
	return postgres.ExecOrWrap(ctx, r.db, queryUpsert, "upsert project member", projectID, userID, role)
}

// Delete removes a user from a project.
func (r *Repository) Delete(ctx context.Context, projectID string, userID string) error {
	return postgres.ExecRequireRowOrWrap(ctx, r.db, queryDelete, "delete project member", projectID, userID)
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

// SetLearningModeOverride sets per-member learning mode override (nullable).
func (r *Repository) SetLearningModeOverride(ctx context.Context, projectID string, userID string, enabled *bool) error {
	return postgres.ExecRequireRowOrWrap(ctx, r.db, querySetLearningModeOverride, "set learning mode override", projectID, userID, enabled)
}

// GetLearningModeOverride returns per-member learning mode override (nullable).
func (r *Repository) GetLearningModeOverride(ctx context.Context, projectID string, userID string) (*bool, bool, error) {
	var v sql.NullBool
	err := r.db.QueryRowContext(ctx, queryGetLearningModeOverride, projectID, userID).Scan(&v)
	if err == nil {
		if !v.Valid {
			return nil, true, nil
		}
		val := v.Bool
		return &val, true, nil
	}
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("get learning mode override: %w", err)
}
