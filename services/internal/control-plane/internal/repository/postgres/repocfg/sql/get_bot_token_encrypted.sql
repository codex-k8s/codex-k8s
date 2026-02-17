-- name: repocfg__get_bot_token_encrypted :one
SELECT bot_token_encrypted
FROM repositories
WHERE id = $1::uuid
LIMIT 1;

