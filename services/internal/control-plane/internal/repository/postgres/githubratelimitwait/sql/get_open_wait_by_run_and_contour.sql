-- name: githubratelimitwait__get_open_wait_by_run_and_contour :one
SELECT *
FROM github_rate_limit_waits
WHERE run_id = $1::uuid
  AND contour_kind = $2
  AND state IN ('open', 'auto_resume_scheduled', 'auto_resume_in_progress', 'manual_action_required')
ORDER BY updated_at DESC, created_at DESC
LIMIT 1;
