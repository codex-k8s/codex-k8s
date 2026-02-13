package dbmodel

import "database/sql"

// RunRow mirrors one agent run row selected from PostgreSQL.
type RunRow struct {
	ID            string
	CorrelationID string
	ProjectID     sql.NullString
	Status        string
	RunPayload    []byte
}
