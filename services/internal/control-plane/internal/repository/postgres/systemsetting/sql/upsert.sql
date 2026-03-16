-- name: systemsetting__upsert :one
INSERT INTO system_settings (
    key,
    value_kind,
    value_json,
    source,
    version,
    updated_by_user_id,
    updated_by_email,
    updated_at
)
VALUES (
    $1,
    $2,
    $3::jsonb,
    $4,
    $5,
    NULLIF($6, '')::uuid,
    NULLIF($7, ''),
    NOW()
)
ON CONFLICT (key) DO UPDATE
SET value_kind = EXCLUDED.value_kind,
    value_json = EXCLUDED.value_json,
    source = EXCLUDED.source,
    version = EXCLUDED.version,
    updated_by_user_id = EXCLUDED.updated_by_user_id,
    updated_by_email = EXCLUDED.updated_by_email,
    updated_at = NOW()
RETURNING
    key,
    value_kind,
    value_json,
    source,
    version,
    updated_by_user_id::text AS updated_by_user_id,
    updated_by_email,
    updated_at;
