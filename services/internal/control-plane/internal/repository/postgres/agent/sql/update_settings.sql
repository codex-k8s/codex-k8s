-- name: agent__update_settings :one
UPDATE agents
SET
    settings = $2::jsonb,
    settings_version = settings_version + 1,
    updated_at = NOW()
WHERE id = $1
  AND settings_version = $3
RETURNING
    id,
    agent_key,
    role_kind,
    project_id::text,
    name,
    is_active,
    settings,
    settings_version;

