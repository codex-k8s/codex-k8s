-- name: configentry__get_by_id :one
SELECT
    id,
    scope,
    kind,
    COALESCE(project_id::text, '') AS project_id,
    COALESCE(repository_id::text, '') AS repository_id,
    key,
    CASE WHEN kind = 'variable' THEN value_plain ELSE '' END AS value,
    sync_targets,
    mutability,
    is_dangerous,
    updated_at::text AS updated_at
FROM config_entries
WHERE id = $1::uuid;

