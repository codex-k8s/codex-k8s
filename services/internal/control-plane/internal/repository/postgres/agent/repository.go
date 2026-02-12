package agent

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agent"
)

var (
	//go:embed sql/find_effective_by_key.sql
	queryFindEffectiveByKey string
)

// Repository stores agent profiles in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL agent profile repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindEffectiveByKey resolves active agent profile by key with project override priority.
func (r *Repository) FindEffectiveByKey(ctx context.Context, projectID string, agentKey string) (domainrepo.Agent, bool, error) {
	var (
		item           domainrepo.Agent
		projectIDValue sql.NullString
	)

	err := r.db.QueryRowContext(ctx, queryFindEffectiveByKey, agentKey, projectID).Scan(
		&item.ID,
		&item.AgentKey,
		&item.RoleKind,
		&projectIDValue,
		&item.Name,
	)
	if err == nil {
		if projectIDValue.Valid {
			item.ProjectID = projectIDValue.String
		}
		return item, true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domainrepo.Agent{}, false, nil
	}
	return domainrepo.Agent{}, false, fmt.Errorf("find effective agent by key: %w", err)
}
