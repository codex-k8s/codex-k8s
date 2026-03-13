-- name: agentrun__clear_wait_context_if_matches :exec
UPDATE agent_runs
SET
    wait_reason = NULL,
    wait_target_kind = NULL,
    wait_target_ref = NULL,
    wait_deadline_at = NULL,
    updated_at = NOW()
WHERE id = $1::uuid
  AND wait_reason = $2
  AND wait_target_kind = $3
  AND wait_target_ref = $4;
