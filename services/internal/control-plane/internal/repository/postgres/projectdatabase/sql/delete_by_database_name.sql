-- name: projectdatabase__delete_by_database_name :one
WITH deleted AS (
    DELETE FROM project_databases
    WHERE database_name = $1
    RETURNING 1
)
SELECT EXISTS (SELECT 1 FROM deleted);
