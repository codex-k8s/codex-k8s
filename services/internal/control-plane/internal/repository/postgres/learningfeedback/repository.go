package learningfeedback

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/learningfeedback"
)

var (
	//go:embed sql/list_for_run.sql
	queryListForRun string
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

// ListForRun returns feedback entries for a run.
func (r *Repository) ListForRun(ctx context.Context, runID string, limit int) ([]domainrepo.Feedback, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, queryListForRun, runID, limit)
	if err != nil {
		return nil, fmt.Errorf("list learning feedback: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domainrepo.Feedback, 0, limit)
	for rows.Next() {
		var item domainrepo.Feedback
		if err := rows.Scan(&item.ID, &item.RunID, &item.RepositoryID, &item.PRNumber, &item.FilePath, &item.Line, &item.Kind, &item.Explanation, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan learning feedback: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate learning feedback: %w", err)
	}
	return out, nil
}

// Insert stores a new feedback record and returns its id.
func (r *Repository) Insert(ctx context.Context, params domainrepo.InsertParams) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(
		ctx,
		queryInsert,
		params.RunID,
		params.RepositoryID,
		params.PRNumber,
		params.FilePath,
		params.Line,
		params.Kind,
		params.Explanation,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert learning feedback: %w", err)
	}
	return id, nil
}
