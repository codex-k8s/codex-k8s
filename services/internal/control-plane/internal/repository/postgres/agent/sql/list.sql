-- name: agent__list :many
SELECT
    id,
    agent_key,
    role_kind,
    project_id::text,
    name,
    is_active,
    settings,
    settings_version
FROM agents
WHERE is_active = TRUE
  AND (
    $3
    OR project_id IS NULL
    OR project_id::text = ANY($1::text[])
  )
ORDER BY
    CASE WHEN project_id IS NULL THEN 0 ELSE 1 END,
    updated_at DESC
LIMIT $2;
