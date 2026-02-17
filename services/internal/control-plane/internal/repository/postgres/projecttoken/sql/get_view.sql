-- name: projecttoken__get_view :one
SELECT
    project_id,
    (platform_token_encrypted IS NOT NULL AND length(platform_token_encrypted) > 0) AS has_platform_token,
    (bot_token_encrypted IS NOT NULL AND length(bot_token_encrypted) > 0) AS has_bot_token,
    bot_username,
    bot_email
FROM project_github_tokens
WHERE project_id = $1::uuid
LIMIT 1;

