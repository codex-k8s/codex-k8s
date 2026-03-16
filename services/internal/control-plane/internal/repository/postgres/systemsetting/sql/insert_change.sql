-- name: systemsetting__insert_change :exec
INSERT INTO system_setting_changes (
    setting_key,
    value_kind,
    value_json,
    previous_value_json,
    source,
    version,
    change_kind,
    actor_user_id,
    actor_email,
    created_at
)
VALUES (
    $1,
    $2,
    $3::jsonb,
    $4::jsonb,
    $5,
    $6,
    $7,
    NULLIF($8, '')::uuid,
    NULLIF($9, ''),
    NOW()
);
