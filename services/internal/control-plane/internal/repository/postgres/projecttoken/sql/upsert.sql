-- name: projecttoken__upsert :exec
INSERT INTO project_github_tokens (
    project_id,
    platform_token_encrypted,
    bot_token_encrypted,
    bot_username,
    bot_email,
    created_at,
    updated_at
)
VALUES (
    $1::uuid,
    $2,
    $3,
    $4,
    $5,
    NOW(),
    NOW()
)
ON CONFLICT (project_id) DO UPDATE
SET platform_token_encrypted = EXCLUDED.platform_token_encrypted,
    bot_token_encrypted = EXCLUDED.bot_token_encrypted,
    bot_username = EXCLUDED.bot_username,
    bot_email = EXCLUDED.bot_email,
    updated_at = NOW();

