-- name: runtimedeploytask__renew_lease :one
UPDATE runtime_deploy_tasks
SET
    lease_until = NOW() + ($3::text)::interval,
    updated_at = NOW()
WHERE run_id = $1::uuid
  AND status = 'running'
  AND lease_owner = $2
RETURNING run_id::text;
