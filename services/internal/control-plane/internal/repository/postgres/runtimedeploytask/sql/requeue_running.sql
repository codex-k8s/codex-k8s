-- name: runtimedeploytask__requeue_running :one
UPDATE runtime_deploy_tasks
SET
    status = 'pending',
    lease_owner = NULL,
    lease_until = NULL,
    last_error = $3,
    finished_at = NULL,
    updated_at = NOW()
WHERE run_id = $1::uuid
  AND status = 'running'
  AND lease_owner = $2
RETURNING run_id::text;
