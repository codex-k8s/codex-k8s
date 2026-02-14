-- name: runtimedeploytask__get_by_run_id :one
SELECT
    run_id::text AS run_id,
    runtime_mode,
    namespace,
    target_env,
    slot_no,
    repository_full_name,
    services_yaml_path,
    build_ref,
    deploy_only,
    status,
    COALESCE(lease_owner, '') AS lease_owner,
    lease_until,
    attempts,
    COALESCE(last_error, '') AS last_error,
    COALESCE(result_namespace, '') AS result_namespace,
    COALESCE(result_target_env, '') AS result_target_env,
    created_at,
    updated_at,
    started_at,
    finished_at
FROM runtime_deploy_tasks
WHERE run_id = $1::uuid
LIMIT 1;
