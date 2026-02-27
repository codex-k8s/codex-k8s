-- name: prompttemplate__insert_version :one
INSERT INTO prompt_templates (
    scope_type,
    scope_id,
    role_key,
    template_kind,
    locale,
    body_markdown,
    source,
    version,
    is_active,
    status,
    checksum,
    change_reason,
    supersedes_version,
    updated_by,
    updated_at,
    activated_at,
    metadata,
    created_at
)
VALUES (
    $1,
    NULLIF($2, '')::uuid,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    NULLIF($12, ''),
    $13,
    $14,
    NOW(),
    $15,
    '{}'::jsonb,
    NOW()
)
RETURNING
    -- Use a single canonical template key for transport/audit DTO.
    CASE
        WHEN scope_type = 'project'
            THEN 'project/' || scope_id::text || '/' || role_key || '/' || template_kind || '/' || locale
        ELSE
            'global/' || role_key || '/' || template_kind || '/' || locale
    END AS template_key,
    version,
    status,
    source,
    checksum,
    body_markdown,
    change_reason,
    supersedes_version,
    updated_by,
    updated_at,
    activated_at;
