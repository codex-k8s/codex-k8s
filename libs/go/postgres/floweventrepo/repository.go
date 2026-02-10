package floweventrepo

import (
	"context"
	"database/sql"
	_ "embed"

	flowdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
)

var (
	//go:embed sql/insert.sql
	queryInsert string
)

type InsertParams = flowdomain.InsertParams

// Repository stores flow events in PostgreSQL.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs a flow event repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Insert appends a flow event row.
func (r *Repository) Insert(ctx context.Context, params InsertParams) error {
	return postgres.InsertFlowEvent(
		ctx,
		r.db,
		queryInsert,
		params.CorrelationID,
		params.ActorType,
		params.ActorID,
		params.EventType,
		[]byte(params.Payload),
		params.CreatedAt,
	)
}
