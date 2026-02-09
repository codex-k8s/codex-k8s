-- name: runqueue__list_running :many
SELECT id, correlation_id, project_id, started_at
FROM agent_runs
WHERE status = 'running'
  AND project_id IS NOT NULL
ORDER BY started_at NULLS FIRST, created_at ASC
LIMIT $1;
