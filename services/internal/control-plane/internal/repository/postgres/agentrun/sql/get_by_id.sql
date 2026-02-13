-- name: agentrun__get_by_id :one
SELECT id, correlation_id, project_id::text AS project_id, status, run_payload
FROM agent_runs
WHERE id = $1
LIMIT 1;
