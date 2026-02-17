-- name: projecttoken__get_encrypted :one
SELECT
    COALESCE(platform_token_encrypted, ''::bytea) AS platform_token_encrypted,
    COALESCE(bot_token_encrypted, ''::bytea) AS bot_token_encrypted,
    bot_username,
    bot_email
FROM project_github_tokens
WHERE project_id = $1::uuid
LIMIT 1;

