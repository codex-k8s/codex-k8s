-- name: githubratelimitwait__clear_session_backpressure :exec
UPDATE agent_sessions
SET
    wait_state = NULL,
    timeout_guard_disabled = false,
    updated_at = NOW()
WHERE run_id = $1::uuid
  AND (wait_state IS NULL OR wait_state = '' OR wait_state = $2);
