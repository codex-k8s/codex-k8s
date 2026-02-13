package dbmodel

import "github.com/jackc/pgx/v5/pgtype"

// RunRow mirrors one agent run row selected from PostgreSQL.
type RunRow struct {
	ID            string      `db:"id"`
	CorrelationID string      `db:"correlation_id"`
	ProjectID     pgtype.Text `db:"project_id"`
	Status        string      `db:"status"`
	RunPayload    []byte      `db:"run_payload"`
}
