-- name: githubratelimitwait__set_run_wait_context :exec
UPDATE agent_runs
SET
    wait_reason = $2,
    wait_target_kind = $3,
    wait_target_ref = $4,
    wait_deadline_at = $5,
    updated_at = NOW()
WHERE id = $1::uuid
  AND (wait_reason IS NULL OR wait_reason = '' OR wait_reason = $2)
  AND (wait_target_kind IS NULL OR wait_target_kind = '' OR wait_target_kind = $3);
