-- name: configentry__delete :exec
DELETE FROM config_entries
WHERE id = $1::uuid;

