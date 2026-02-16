-- name: configentry__list :many
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
WHERE scope = $1
  AND (CASE WHEN $2::text = '' THEN project_id IS NULL ELSE project_id = $2::uuid END)
  AND (CASE WHEN $3::text = '' THEN repository_id IS NULL ELSE repository_id = $3::uuid END)
ORDER BY scope ASC, key ASC
LIMIT $4;

