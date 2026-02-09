-- name: user__get_by_email :one
SELECT id, email, COALESCE(github_user_id, 0) AS github_user_id, COALESCE(github_login, '') AS github_login, is_platform_admin
FROM users
WHERE email = LOWER($1);

