-- name: user__get_by_github_login :one
SELECT id, email, COALESCE(github_user_id, 0) AS github_user_id, COALESCE(github_login, '') AS github_login, is_platform_admin
FROM users
WHERE github_login IS NOT NULL
  AND github_login <> ''
  AND LOWER(github_login) = LOWER($1)
LIMIT 1;

