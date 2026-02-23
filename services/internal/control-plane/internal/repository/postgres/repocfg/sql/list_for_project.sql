-- name: repocfg__list_for_project :many
SELECT
    id,
    project_id,
    alias,
    role,
    default_ref,
    provider,
    external_id,
    owner,
    name,
    services_yaml_path,
    COALESCE(docs_root_path, '') AS docs_root_path,
    bot_username,
    bot_email,
    COALESCE(preflight_updated_at::text, '') AS preflight_updated_at
FROM repositories
WHERE project_id = $1::uuid
ORDER BY alias ASC, provider ASC, owner ASC, name ASC
LIMIT $2;
