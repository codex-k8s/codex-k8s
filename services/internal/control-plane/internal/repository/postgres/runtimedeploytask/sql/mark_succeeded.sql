-- name: runtimedeploytask__mark_succeeded :one
UPDATE runtime_deploy_tasks
SET
    status = 'succeeded',
    lease_owner = NULL,
    lease_until = NULL,
    last_error = NULL,
    result_namespace = $3,
    result_target_env = $4,
    finished_at = NOW(),
    updated_at = NOW()
WHERE run_id = $1::uuid
  AND status = 'running'
  AND lease_owner = $2
RETURNING run_id::text;
