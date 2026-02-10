-- name: project__get_learning_mode_default :one
SELECT COALESCE((settings->>'learning_mode_default')::boolean, false) AS learning_mode_default
FROM projects
WHERE id = $1::uuid;

