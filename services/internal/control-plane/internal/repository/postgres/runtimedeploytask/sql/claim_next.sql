-- name: runtimedeploytask__claim_next :one
WITH candidate AS (
    SELECT run_id
    FROM runtime_deploy_tasks
    WHERE status = 'pending'
       OR (
           status = 'running'
           AND lease_until IS NOT NULL
           AND lease_until < NOW()
       )
    ORDER BY updated_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
UPDATE runtime_deploy_tasks t
SET
    status = 'running',
    lease_owner = $1,
    lease_until = NOW() + ($2::text)::interval,
    attempts = t.attempts + 1,
    last_error = NULL,
    started_at = COALESCE(t.started_at, NOW()),
    finished_at = NULL,
    updated_at = NOW()
FROM candidate
WHERE t.run_id = candidate.run_id
RETURNING
    t.run_id::text AS run_id,
    t.runtime_mode,
    t.namespace,
    t.target_env,
    t.slot_no,
    t.repository_full_name,
    t.services_yaml_path,
    t.build_ref,
    t.deploy_only,
    t.status,
    COALESCE(t.lease_owner, '') AS lease_owner,
    t.lease_until,
    t.attempts,
    COALESCE(t.last_error, '') AS last_error,
    COALESCE(t.result_namespace, '') AS result_namespace,
    COALESCE(t.result_target_env, '') AS result_target_env,
    t.created_at,
    t.updated_at,
    t.started_at,
    t.finished_at,
    COALESCE(t.logs_json, '[]'::jsonb) AS logs_json;
