-- name: staffrun__list_for_user :many
SELECT
    ar.id,
    ar.correlation_id,
    ar.project_id::text AS project_id,
    COALESCE(p.slug, '') AS project_slug,
    COALESCE(p.name, '') AS project_name,
    ar.status,
    ar.created_at,
    ar.started_at,
    ar.finished_at
FROM agent_runs ar
JOIN project_members pm ON pm.project_id = ar.project_id
JOIN projects p ON p.id = ar.project_id
WHERE pm.user_id = $1::uuid
  AND ar.project_id IS NOT NULL
ORDER BY ar.created_at DESC
LIMIT $2;
