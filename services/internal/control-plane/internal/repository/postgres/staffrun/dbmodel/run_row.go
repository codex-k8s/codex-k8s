package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// RunRow mirrors one staff run row selected from PostgreSQL.
type RunRow struct {
	ID            string             `db:"id"`
	CorrelationID string             `db:"correlation_id"`
	ProjectID     pgtype.Text        `db:"project_id"`
	ProjectSlug   string             `db:"project_slug"`
	ProjectName   string             `db:"project_name"`
	IssueNumber   pgtype.Int4        `db:"issue_number"`
	IssueURL      pgtype.Text        `db:"issue_url"`
	TriggerKind   pgtype.Text        `db:"trigger_kind"`
	TriggerLabel  pgtype.Text        `db:"trigger_label"`
	JobName       pgtype.Text        `db:"job_name"`
	JobNamespace  pgtype.Text        `db:"job_namespace"`
	Namespace     pgtype.Text        `db:"namespace"`
	WaitState     pgtype.Text        `db:"wait_state"`
	WaitReason    pgtype.Text        `db:"wait_reason"`
	PRURL         pgtype.Text        `db:"pr_url"`
	PRNumber      pgtype.Int4        `db:"pr_number"`
	Status        string             `db:"status"`
	CreatedAt     time.Time          `db:"created_at"`
	StartedAt     pgtype.Timestamptz `db:"started_at"`
	FinishedAt    pgtype.Timestamptz `db:"finished_at"`
}
