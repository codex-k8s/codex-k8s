-- name: githubratelimitwait__list_waits_by_run_id :many
SELECT *
FROM github_rate_limit_waits
WHERE run_id = $1::uuid
ORDER BY updated_at DESC, created_at DESC, id DESC;
