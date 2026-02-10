-- name: user__ensure_owner :one
WITH clear_prev_owner AS (
    UPDATE users
    SET is_platform_owner = false,
        updated_at = NOW()
    WHERE is_platform_owner = true
      AND email <> LOWER($1)
)
INSERT INTO users (email, is_platform_admin, is_platform_owner)
VALUES (LOWER($1), TRUE, TRUE)
ON CONFLICT (email) DO UPDATE
SET is_platform_admin = TRUE,
    is_platform_owner = TRUE,
    updated_at = NOW()
RETURNING id, email, COALESCE(github_user_id, 0) AS github_user_id, COALESCE(github_login, '') AS github_login, is_platform_admin, is_platform_owner;
