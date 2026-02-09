-- name: runqueue__mark_run_running :exec
UPDATE agent_runs
SET status = 'running',
    project_id = $2::uuid,
    started_at = COALESCE(started_at, NOW()),
    updated_at = NOW()
WHERE id = $1
  AND status = 'pending';
