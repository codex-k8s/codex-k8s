-- name: staffrun__get_by_id :one
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
LEFT JOIN projects p ON p.id = ar.project_id
WHERE ar.id = $1::uuid;

