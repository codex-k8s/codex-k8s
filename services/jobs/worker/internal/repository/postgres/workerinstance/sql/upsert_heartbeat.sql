-- name: workerinstance__upsert_heartbeat :exec
INSERT INTO worker_instances (
    worker_id,
    namespace,
    pod_name,
    status,
    started_at,
    heartbeat_at,
    expires_at,
    created_at,
    updated_at
)
VALUES (
    $1,
    $2,
    $3,
    'active',
    $4,
    $5,
    $6,
    NOW(),
    NOW()
)
ON CONFLICT (worker_id) DO UPDATE
SET namespace = EXCLUDED.namespace,
    pod_name = EXCLUDED.pod_name,
    status = 'active',
    started_at = EXCLUDED.started_at,
    heartbeat_at = EXCLUDED.heartbeat_at,
    expires_at = EXCLUDED.expires_at,
    updated_at = NOW();
