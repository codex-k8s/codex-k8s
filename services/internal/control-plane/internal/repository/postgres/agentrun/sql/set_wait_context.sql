-- name: agentrun__set_wait_context :exec
UPDATE agent_runs
SET
    wait_reason = NULLIF($2, ''),
    wait_target_kind = NULLIF($3, ''),
    wait_target_ref = NULLIF($4, ''),
    wait_deadline_at = $5,
    updated_at = NOW()
WHERE id = $1::uuid;
