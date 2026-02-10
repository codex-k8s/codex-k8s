-- name: project__delete_by_id :exec
DELETE FROM projects
WHERE id = $1::uuid;

