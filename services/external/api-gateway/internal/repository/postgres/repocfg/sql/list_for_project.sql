-- name: repocfg__list_for_project :many
SELECT id, project_id, provider, external_id, owner, name, services_yaml_path
FROM repositories
WHERE project_id = $1::uuid
ORDER BY provider ASC, owner ASC, name ASC
LIMIT $2;

