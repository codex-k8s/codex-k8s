-- name: githubratelimitwait__clear_run_wait_context :exec
UPDATE agent_runs
SET
    wait_reason = NULL,
    wait_target_kind = NULL,
    wait_target_ref = NULL,
    wait_deadline_at = NULL,
    updated_at = NOW()
WHERE id = $1::uuid
  AND (wait_reason IS NULL OR wait_reason = '' OR wait_reason = $2)
  AND (wait_target_kind IS NULL OR wait_target_kind = '' OR wait_target_kind = $3);
