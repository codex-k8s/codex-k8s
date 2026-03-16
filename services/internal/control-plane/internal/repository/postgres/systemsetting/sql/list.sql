-- name: systemsetting__list :many
SELECT
    key,
    value_kind,
    value_json,
    source,
    version,
    updated_by_user_id::text AS updated_by_user_id,
    updated_by_email,
    updated_at
FROM system_settings
ORDER BY key ASC;
