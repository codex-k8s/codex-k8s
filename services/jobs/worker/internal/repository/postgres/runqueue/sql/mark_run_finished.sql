-- name: runqueue__mark_run_finished :exec
UPDATE agent_runs
SET status = $2,
    finished_at = $3,
    lease_owner = NULL,
    lease_until = NULL,
    updated_at = NOW()
WHERE id = $1
  AND lease_owner = $4
  AND status = 'running';
