-- name: agentsession__set_wait_state_by_run_id :exec
UPDATE agent_sessions
SET
    wait_state = $2,
    timeout_guard_disabled = $3,
    last_heartbeat_at = $4,
    updated_at = NOW()
WHERE run_id = $1;
