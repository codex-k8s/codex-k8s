-- name: agentrun__get_run_id_by_correlation_id :one
SELECT id
FROM agent_runs
WHERE correlation_id = $1;
