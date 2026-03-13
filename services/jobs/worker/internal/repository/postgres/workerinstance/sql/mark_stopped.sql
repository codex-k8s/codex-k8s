-- name: workerinstance__mark_stopped :exec
UPDATE worker_instances
SET status = 'stopped',
    heartbeat_at = $2,
    expires_at = $2,
    updated_at = NOW()
WHERE worker_id = $1;
