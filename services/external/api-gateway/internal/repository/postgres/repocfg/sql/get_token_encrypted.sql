-- name: repocfg__get_token_encrypted :one
SELECT token_encrypted
FROM repositories
WHERE id = $1::uuid;

