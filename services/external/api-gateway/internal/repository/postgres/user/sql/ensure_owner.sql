-- name: user__ensure_owner :one
INSERT INTO users (email, is_platform_admin)
VALUES (LOWER($1), TRUE)
ON CONFLICT (email) DO UPDATE
SET is_platform_admin = TRUE,
    updated_at = NOW()
RETURNING id, email, github_user_id, github_login, is_platform_admin;

