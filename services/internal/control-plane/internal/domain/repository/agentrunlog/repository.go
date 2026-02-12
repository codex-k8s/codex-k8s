package agentrunlog

import (
	"context"
	"encoding/json"
	"time"
)

// Repository updates and purges run-scoped agent logs in agent_runs.
type Repository interface {
	// UpsertRunAgentLogs stores latest agent execution logs for one run.
	UpsertRunAgentLogs(ctx context.Context, runID string, logs json.RawMessage) error
	// CleanupRunAgentLogsFinishedBefore clears logs for finished runs older than cutoff.
	CleanupRunAgentLogsFinishedBefore(ctx context.Context, finishedBefore time.Time) (int64, error)
}
