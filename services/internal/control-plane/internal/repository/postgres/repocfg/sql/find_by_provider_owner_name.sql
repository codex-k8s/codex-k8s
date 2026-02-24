-- name: repocfg__find_by_provider_owner_name :one
SELECT project_id, id, services_yaml_path, default_ref
FROM repositories
WHERE provider = $1
  AND lower(owner) = lower($2)
  AND lower(name) = lower($3)
LIMIT 1;
