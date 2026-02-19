UPDATE run_access_keys
SET
    status = 'revoked',
    revoked_at = $2,
    updated_at = $3
WHERE run_id = $1
RETURNING
    run_id,
    project_id::text AS project_id,
    correlation_id,
    runtime_mode,
    namespace,
    target_env,
    key_hash,
    status,
    issued_at,
    expires_at,
    revoked_at,
    last_used_at,
    created_by,
    created_at,
    updated_at;
