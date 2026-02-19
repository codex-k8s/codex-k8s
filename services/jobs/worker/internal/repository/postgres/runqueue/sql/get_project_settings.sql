-- name: runqueue__get_project_settings :one
SELECT settings
FROM projects
WHERE id = $1::uuid;
