package dbmodel

import (
	"database/sql"
	"time"
)

// RunRow mirrors one staff run row selected from PostgreSQL.
type RunRow struct {
	ID            string
	CorrelationID string
	ProjectID     sql.NullString
	ProjectSlug   string
	ProjectName   string
	IssueNumber   sql.NullInt32
	IssueURL      sql.NullString
	TriggerKind   sql.NullString
	TriggerLabel  sql.NullString
	JobName       sql.NullString
	JobNamespace  sql.NullString
	Namespace     sql.NullString
	PRURL         sql.NullString
	PRNumber      sql.NullInt32
	Status        string
	CreatedAt     time.Time
	StartedAt     sql.NullTime
	FinishedAt    sql.NullTime
}
