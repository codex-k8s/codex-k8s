-- name: projectdatabase__get_by_database_name :one
SELECT
    project_id,
    environment,
    database_name,
    created_at,
    updated_at
FROM project_databases
WHERE database_name = $1
LIMIT 1;
