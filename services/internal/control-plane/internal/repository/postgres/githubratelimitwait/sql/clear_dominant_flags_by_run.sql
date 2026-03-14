-- name: githubratelimitwait__clear_dominant_flags_by_run :exec
UPDATE github_rate_limit_waits
SET dominant_for_run = false
WHERE run_id = $1::uuid
  AND dominant_for_run = true;
