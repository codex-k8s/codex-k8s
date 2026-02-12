SELECT id, project_id, provider, external_id, owner, name, services_yaml_path
FROM repositories
WHERE id = $1
LIMIT 1;
