package agentrun

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/google/uuid"

	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
)

var (
	//go:embed sql/create_pending_if_absent.sql
	queryCreatePendingIfAbsent string
	//go:embed sql/get_run_id_by_correlation_id.sql
	queryGetRunIDByCorrelationID string
)

// Repository stores agent runs in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs PostgreSQL agent run repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreatePendingIfAbsent inserts pending run or returns an existing run id.
func (r *Repository) CreatePendingIfAbsent(ctx context.Context, params domainrepo.CreateParams) (domainrepo.CreateResult, error) {
	runID := uuid.NewString()
	var insertedRunID string
	err := r.db.QueryRowContext(
		ctx,
		queryCreatePendingIfAbsent,
		runID,
		params.CorrelationID,
		params.ProjectID,
		[]byte(params.RunPayload),
		params.LearningMode,
	).Scan(&insertedRunID)
	if err == nil {
		return domainrepo.CreateResult{
			RunID:    insertedRunID,
			Inserted: true,
		}, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return domainrepo.CreateResult{}, fmt.Errorf("insert agent run: %w", err)
	}

	var existingRunID string
	if err := r.db.QueryRowContext(
		ctx,
		queryGetRunIDByCorrelationID,
		params.CorrelationID,
	).Scan(&existingRunID); err != nil {
		return domainrepo.CreateResult{}, fmt.Errorf("get existing agent run by correlation id: %w", err)
	}

	return domainrepo.CreateResult{
		RunID:    existingRunID,
		Inserted: false,
	}, nil
}
