-- name: runtimedeploytask__mark_failed :one
UPDATE runtime_deploy_tasks
SET
    status = 'failed',
    lease_owner = NULL,
    lease_until = NULL,
    last_error = $3,
    finished_at = NOW(),
    updated_at = NOW()
WHERE run_id = $1::uuid
  AND status = 'running'
  AND lease_owner = $2
RETURNING run_id::text;
