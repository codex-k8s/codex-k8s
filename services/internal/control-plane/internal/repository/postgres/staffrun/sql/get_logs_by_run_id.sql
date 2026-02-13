-- name: staffrun__get_logs_by_run_id :one
SELECT
    ar.id AS run_id,
    ar.status,
    ar.updated_at,
    COALESCE(ar.agent_logs_json, '{}'::jsonb) AS snapshot_json
FROM agent_runs ar
WHERE ar.id = $1::uuid
LIMIT 1;
