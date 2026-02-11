-- name: repocfg__find_by_provider_external_id :one
SELECT project_id, id, services_yaml_path
FROM repositories
WHERE provider = $1
  AND external_id = $2
LIMIT 1;

