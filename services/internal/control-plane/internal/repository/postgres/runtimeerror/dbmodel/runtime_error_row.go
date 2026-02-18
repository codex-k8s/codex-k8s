package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// RuntimeErrorRow mirrors one runtime_errors row selected from PostgreSQL.
type RuntimeErrorRow struct {
	ID            string             `db:"id"`
	Source        string             `db:"source"`
	Level         string             `db:"level"`
	Message       string             `db:"message"`
	DetailsJSON   []byte             `db:"details_json"`
	StackTrace    pgtype.Text        `db:"stack_trace"`
	CorrelationID pgtype.Text        `db:"correlation_id"`
	RunID         pgtype.Text        `db:"run_id"`
	ProjectID     pgtype.Text        `db:"project_id"`
	Namespace     pgtype.Text        `db:"namespace"`
	JobName       pgtype.Text        `db:"job_name"`
	ViewedAt      pgtype.Timestamptz `db:"viewed_at"`
	ViewedBy      pgtype.Text        `db:"viewed_by"`
	CreatedAt     time.Time          `db:"created_at"`
}
