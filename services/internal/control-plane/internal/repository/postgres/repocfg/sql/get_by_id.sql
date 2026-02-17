-- name: repocfg__get_by_id :one
SELECT
    id,
    project_id,
    provider,
    external_id,
    owner,
    name,
    services_yaml_path,
    bot_username,
    bot_email,
    COALESCE(preflight_updated_at::text, '') AS preflight_updated_at
FROM repositories
WHERE id = $1
LIMIT 1;
