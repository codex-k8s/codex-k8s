-- name: githubratelimitwait__get_wait_by_signal_id :one
SELECT *
FROM github_rate_limit_waits
WHERE signal_id = $1;
