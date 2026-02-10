package flowevent

import (
	"context"
	"database/sql"

	libflow "github.com/codex-k8s/codex-k8s/libs/go/postgres/floweventrepo"
	domainrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
)

// Repository stores flow events in PostgreSQL.
type Repository struct {
	inner *libflow.Repository
}

// NewRepository constructs PostgreSQL flow event repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{inner: libflow.NewRepository(db)}
}

// Insert appends a flow event row to PostgreSQL.
func (r *Repository) Insert(ctx context.Context, params domainrepo.InsertParams) error {
	return r.inner.Insert(ctx, libflow.InsertParams{
		CorrelationID: params.CorrelationID,
		ActorType:     params.ActorType,
		ActorID:       params.ActorID,
		EventType:     params.EventType,
		Payload:       []byte(params.Payload),
		CreatedAt:     params.CreatedAt,
	})
}
