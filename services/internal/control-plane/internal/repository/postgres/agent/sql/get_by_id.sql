-- name: agent__get_by_id :one
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
WHERE id = $1
LIMIT 1;

