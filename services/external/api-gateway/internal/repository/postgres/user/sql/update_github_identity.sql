-- name: user__update_github_identity :exec
UPDATE users
SET github_user_id = $2,
    github_login = $3,
    updated_at = NOW()
WHERE id = $1::uuid;

