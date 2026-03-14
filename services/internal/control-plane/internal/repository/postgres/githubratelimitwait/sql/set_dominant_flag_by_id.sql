-- name: githubratelimitwait__set_dominant_flag_by_id :exec
UPDATE github_rate_limit_waits
SET dominant_for_run = true
WHERE id = $1::uuid;
