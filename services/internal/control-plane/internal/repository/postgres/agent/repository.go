package agent

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agent"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/find_effective_by_key.sql
	queryFindEffectiveByKey string
)

// Repository stores agent profiles in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs PostgreSQL agent profile repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindEffectiveByKey resolves active agent profile by key with project override priority.
func (r *Repository) FindEffectiveByKey(ctx context.Context, projectID string, agentKey string) (domainrepo.Agent, bool, error) {
	var (
		item           domainrepo.Agent
		projectIDValue pgtype.Text
	)

	err := r.db.QueryRow(ctx, queryFindEffectiveByKey, agentKey, projectID).Scan(
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
	if errors.Is(err, pgx.ErrNoRows) {
		return domainrepo.Agent{}, false, nil
	}
	return domainrepo.Agent{}, false, fmt.Errorf("find effective agent by key: %w", err)
}
