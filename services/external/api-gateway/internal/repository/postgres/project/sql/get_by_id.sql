-- name: project__get_by_id :one
SELECT id, slug, name
FROM projects
WHERE id = $1::uuid;

