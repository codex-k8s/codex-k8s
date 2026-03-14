-- name: githubratelimitwait__list_open_waits_by_run_for_update :many
SELECT *
FROM github_rate_limit_waits
WHERE run_id = $1::uuid
  AND state IN ('open', 'auto_resume_scheduled', 'auto_resume_in_progress', 'manual_action_required')
ORDER BY updated_at DESC, created_at DESC, id DESC
FOR UPDATE;
