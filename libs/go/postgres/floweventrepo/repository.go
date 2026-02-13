package floweventrepo

import (
	"context"
	_ "embed"

	flowdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	//go:embed sql/insert.sql
	queryInsert string
)

type InsertParams = flowdomain.InsertParams

// Repository stores flow events in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a flow event repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Insert appends a flow event row.
func (r *Repository) Insert(ctx context.Context, params InsertParams) error {
	return postgres.InsertFlowEvent(
		ctx,
		r.db,
		queryInsert,
		params.CorrelationID,
		string(params.ActorType),
		string(params.ActorID),
		string(params.EventType),
		[]byte(params.Payload),
		params.CreatedAt,
	)
}
