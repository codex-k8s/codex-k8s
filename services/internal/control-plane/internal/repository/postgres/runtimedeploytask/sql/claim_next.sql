-- name: runtimedeploytask__claim_next :one
WITH candidate AS (
    SELECT t.run_id
    FROM runtime_deploy_tasks t
    WHERE (
        t.status = 'pending'
        OR (
            t.status = 'running'
            AND t.lease_until IS NOT NULL
            AND t.lease_until < NOW()
        )
        OR (
            t.status = 'running'
            AND t.lease_until IS NOT NULL
            AND t.updated_at < NOW() - INTERVAL '2 minutes'
        )
    )
      AND NOT EXISTS (
          SELECT 1
          FROM runtime_deploy_tasks active
          WHERE active.run_id <> t.run_id
            AND active.status = 'running'
            AND active.lease_until IS NOT NULL
            AND active.lease_until >= NOW()
            AND active.namespace = t.namespace
            AND active.target_env = t.target_env
      )
    ORDER BY t.updated_at ASC
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
