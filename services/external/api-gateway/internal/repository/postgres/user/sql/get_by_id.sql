-- name: user__get_by_id :one
SELECT id, email, COALESCE(github_user_id, 0) AS github_user_id, COALESCE(github_login, '') AS github_login, is_platform_admin, is_platform_owner
FROM users
WHERE id = $1::uuid;

