package dbmodel

import "database/sql"

// RunRow mirrors one agent run row selected from PostgreSQL.
type RunRow struct {
	ID            string         `db:"id"`
	CorrelationID string         `db:"correlation_id"`
	ProjectID     sql.NullString `db:"project_id"`
	Status        string         `db:"status"`
	RunPayload    []byte         `db:"run_payload"`
}
