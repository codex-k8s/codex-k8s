-- name: runtimedeploytask__list_recent :many
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
    cancel_requested_at,
    COALESCE(cancel_requested_by, '') AS cancel_requested_by,
    COALESCE(cancel_reason, '') AS cancel_reason,
    stop_requested_at,
    COALESCE(stop_requested_by, '') AS stop_requested_by,
    COALESCE(stop_reason, '') AS stop_reason,
    COALESCE(terminal_status_source, '') AS terminal_status_source,
    terminal_event_seq,
    created_at,
    updated_at,
    started_at,
    finished_at,
    '[]'::jsonb AS logs_json
FROM runtime_deploy_tasks
WHERE ($1::text = '' OR status = $1::text)
  AND ($2::text = '' OR target_env = $2::text)
ORDER BY updated_at DESC
LIMIT $3;
