-- name: user__delete_by_id :exec
DELETE FROM users
WHERE id = $1::uuid;

