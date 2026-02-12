-- name: platformtoken__get :one
SELECT
    platform_token_encrypted,
    bot_token_encrypted
FROM platform_github_tokens
WHERE id = 1;
