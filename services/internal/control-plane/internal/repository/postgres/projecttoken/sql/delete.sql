-- name: projecttoken__delete :exec
DELETE FROM project_github_tokens
WHERE project_id = $1::uuid;

