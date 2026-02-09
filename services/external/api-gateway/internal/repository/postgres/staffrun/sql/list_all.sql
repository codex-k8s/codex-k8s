-- name: staffrun__list_all :many
SELECT
    id,
    correlation_id,
    COALESCE(project_id::text, '') AS project_id,
    status,
    created_at::text,
    COALESCE(started_at::text, '') AS started_at,
    COALESCE(finished_at::text, '') AS finished_at
FROM agent_runs
ORDER BY created_at DESC
LIMIT $1;

