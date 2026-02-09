-- name: user__list_users :many
SELECT id, email, COALESCE(github_user_id, 0) AS github_user_id, COALESCE(github_login, '') AS github_login, is_platform_admin
FROM users
ORDER BY is_platform_admin DESC, email ASC
LIMIT $1;

