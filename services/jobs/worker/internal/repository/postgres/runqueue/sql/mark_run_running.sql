-- name: runqueue__mark_run_running :exec
UPDATE agent_runs
SET status = 'running',
    project_id = $2::uuid,
    lease_owner = $3,
    lease_until = NOW() + ($4::text)::interval,
    started_at = COALESCE(started_at, NOW()),
    updated_at = NOW()
WHERE id = $1
  AND status = 'pending';
