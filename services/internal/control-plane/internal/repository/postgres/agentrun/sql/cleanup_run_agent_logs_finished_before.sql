-- name: agentrun__cleanup_run_agent_logs_finished_before :exec
UPDATE agent_runs
SET agent_logs_json = NULL,
    updated_at = NOW()
WHERE agent_logs_json IS NOT NULL
  AND finished_at IS NOT NULL
  AND finished_at < $1::timestamptz;
