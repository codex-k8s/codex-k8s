-- name: repocfg__upsert_bot_params :exec
UPDATE repositories
SET bot_token_encrypted = $2,
    bot_username = $3,
    bot_email = $4,
    updated_at = NOW()
WHERE id = $1::uuid;

