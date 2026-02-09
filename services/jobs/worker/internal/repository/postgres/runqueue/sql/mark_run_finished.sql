-- name: runqueue__mark_run_finished :exec
UPDATE agent_runs
SET status = $2,
    finished_at = $3,
    updated_at = NOW()
WHERE id = $1
  AND status = 'running';
