-- name: staffrun__get_correlation_by_run_id :one
SELECT correlation_id, COALESCE(project_id::text, '') AS project_id
FROM agent_runs
WHERE id = $1;

