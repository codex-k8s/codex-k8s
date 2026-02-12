-- name: platformtoken__upsert :one
INSERT INTO platform_github_tokens (
    id,
    platform_token_encrypted,
    bot_token_encrypted,
    updated_at
)
VALUES (
    1,
    $1,
    $2,
    NOW()
)
ON CONFLICT (id) DO UPDATE SET
    platform_token_encrypted = EXCLUDED.platform_token_encrypted,
    bot_token_encrypted = EXCLUDED.bot_token_encrypted,
    updated_at = NOW()
RETURNING
    platform_token_encrypted,
    bot_token_encrypted;
