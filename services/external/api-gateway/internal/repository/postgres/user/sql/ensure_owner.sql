-- name: user__ensure_owner :one
INSERT INTO users (email, is_platform_admin)
VALUES (LOWER($1), TRUE)
ON CONFLICT (email) DO UPDATE
SET is_platform_admin = TRUE,
    updated_at = NOW()
RETURNING id, email, COALESCE(github_user_id, 0) AS github_user_id, COALESCE(github_login, '') AS github_login, is_platform_admin;
