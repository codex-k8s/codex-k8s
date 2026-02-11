-- name: runqueue__list_running :many
-- run_payload column is introduced by services/internal/control-plane/cmd/cli/migrations/20260206191000_day1_webhook_ingest.sql.
SELECT id, correlation_id, project_id, learning_mode, run_payload, started_at
FROM agent_runs
WHERE status = 'running'
  AND project_id IS NOT NULL
ORDER BY started_at NULLS FIRST, created_at ASC
LIMIT $1;
