-- name: projectdatabase__upsert :one
INSERT INTO project_databases (
    project_id,
    environment,
    database_name
)
VALUES (
    $1::uuid,
    $2,
    $3
)
ON CONFLICT (database_name)
DO UPDATE SET
    project_id = EXCLUDED.project_id,
    environment = EXCLUDED.environment,
    updated_at = NOW()
RETURNING
    project_id,
    environment,
    database_name,
    created_at,
    updated_at;
