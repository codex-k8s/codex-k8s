-- name: githubratelimitwait__get_wait_by_id :one
SELECT *
FROM github_rate_limit_waits
WHERE id = $1::uuid;
