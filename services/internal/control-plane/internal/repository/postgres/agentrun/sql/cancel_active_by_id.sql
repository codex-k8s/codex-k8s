-- name: agentrun__cancel_active_by_id :exec
UPDATE agent_runs
SET status = 'canceled',
    finished_at = NOW(),
    lease_owner = NULL,
    lease_until = NULL,
    updated_at = NOW()
WHERE id = $1
  AND status IN ('pending', 'running');
