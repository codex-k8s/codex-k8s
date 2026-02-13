package dbmodel

import (
	"database/sql"
	"time"
)

// RunRow mirrors one staff run row selected from PostgreSQL.
type RunRow struct {
	ID            string         `db:"id"`
	CorrelationID string         `db:"correlation_id"`
	ProjectID     sql.NullString `db:"project_id"`
	ProjectSlug   string         `db:"project_slug"`
	ProjectName   string         `db:"project_name"`
	IssueNumber   sql.NullInt32  `db:"issue_number"`
	IssueURL      sql.NullString `db:"issue_url"`
	TriggerKind   sql.NullString `db:"trigger_kind"`
	TriggerLabel  sql.NullString `db:"trigger_label"`
	JobName       sql.NullString `db:"job_name"`
	JobNamespace  sql.NullString `db:"job_namespace"`
	Namespace     sql.NullString `db:"namespace"`
	PRURL         sql.NullString `db:"pr_url"`
	PRNumber      sql.NullInt32  `db:"pr_number"`
	Status        string         `db:"status"`
	CreatedAt     time.Time      `db:"created_at"`
	StartedAt     sql.NullTime   `db:"started_at"`
	FinishedAt    sql.NullTime   `db:"finished_at"`
}
