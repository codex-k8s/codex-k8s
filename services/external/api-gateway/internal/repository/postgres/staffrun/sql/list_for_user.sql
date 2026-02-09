-- name: staffrun__list_for_user :many
SELECT
    ar.id,
    ar.correlation_id,
    ar.project_id::text,
    ar.status,
    ar.created_at::text,
    COALESCE(ar.started_at::text, '') AS started_at,
    COALESCE(ar.finished_at::text, '') AS finished_at
FROM agent_runs ar
JOIN project_members pm ON pm.project_id = ar.project_id
WHERE pm.user_id = $1::uuid
  AND ar.project_id IS NOT NULL
ORDER BY ar.created_at DESC
LIMIT $2;

