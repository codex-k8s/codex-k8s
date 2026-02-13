package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ActionRequestRow mirrors one mcp_action_requests row enriched with run/project metadata.
type ActionRequestRow struct {
	ID            int64       `db:"id"`
	CorrelationID string      `db:"correlation_id"`
	RunID         pgtype.Text `db:"run_id"`
	ProjectID     pgtype.Text `db:"project_id"`
	ProjectSlug   pgtype.Text `db:"project_slug"`
	ProjectName   pgtype.Text `db:"project_name"`
	IssueNumber   pgtype.Int4 `db:"issue_number"`
	PRNumber      pgtype.Int4 `db:"pr_number"`
	TriggerLabel  pgtype.Text `db:"trigger_label"`
	ToolName      string      `db:"tool_name"`
	Action        string      `db:"action"`
	TargetRef     []byte      `db:"target_ref"`
	ApprovalMode  string      `db:"approval_mode"`
	ApprovalState string      `db:"approval_state"`
	RequestedBy   string      `db:"requested_by"`
	AppliedBy     pgtype.Text `db:"applied_by"`
	Payload       []byte      `db:"payload"`
	CreatedAt     time.Time   `db:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at"`
}
