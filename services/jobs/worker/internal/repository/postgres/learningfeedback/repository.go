package learningfeedback

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/learningfeedback"
)

var (
	//go:embed sql/insert.sql
	queryInsert string
)

// Repository stores learning feedback in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL learning feedback repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Insert creates a new learning_feedback record.
func (r *Repository) Insert(ctx context.Context, params domainrepo.InsertParams) error {
	if _, err := r.db.ExecContext(ctx, queryInsert, params.RunID, params.Kind, params.Explanation); err != nil {
		return fmt.Errorf("insert learning feedback: %w", err)
	}
	return nil
}
