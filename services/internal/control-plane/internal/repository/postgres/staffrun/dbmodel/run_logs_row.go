package dbmodel

import "github.com/jackc/pgx/v5/pgtype"

// RunLogsRow mirrors one run logs snapshot row from agent_runs.
type RunLogsRow struct {
	RunID        string             `db:"run_id"`
	Status       string             `db:"status"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at"`
	SnapshotJSON []byte             `db:"snapshot_json"`
}
