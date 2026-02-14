package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// RunLookupRow mirrors one lookup row for self-improve run diagnostics.
type RunLookupRow struct {
	RunID              string             `db:"run_id"`
	CorrelationID      string             `db:"correlation_id"`
	ProjectID          pgtype.Text        `db:"project_id"`
	RepositoryFullName string             `db:"repository_full_name"`
	AgentKey           string             `db:"agent_key"`
	IssueNumber        pgtype.Int8        `db:"issue_number"`
	IssueURL           string             `db:"issue_url"`
	PullRequestNumber  pgtype.Int8        `db:"pull_request_number"`
	PullRequestURL     string             `db:"pull_request_url"`
	TriggerKind        string             `db:"trigger_kind"`
	TriggerLabel       string             `db:"trigger_label"`
	Status             string             `db:"status"`
	CreatedAt          time.Time          `db:"created_at"`
	StartedAt          pgtype.Timestamptz `db:"started_at"`
	FinishedAt         pgtype.Timestamptz `db:"finished_at"`
}
