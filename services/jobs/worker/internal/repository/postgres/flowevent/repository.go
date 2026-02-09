package flowevent

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
)

var (
	//go:embed sql/insert.sql
	queryInsert string
)

// Repository stores flow events in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL flow event repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Insert appends one flow event row.
func (r *Repository) Insert(ctx context.Context, params domainrepo.InsertParams) error {
	_, err := r.db.ExecContext(
		ctx,
		queryInsert,
		params.CorrelationID,
		params.ActorType,
		params.ActorID,
		params.EventType,
		[]byte(params.Payload),
		params.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert flow event: %w", err)
	}
	return nil
}
