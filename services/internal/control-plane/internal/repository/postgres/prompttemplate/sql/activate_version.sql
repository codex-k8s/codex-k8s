-- name: prompttemplate__activate_version :one
UPDATE prompt_templates
SET
    is_active = TRUE,
    status = 'active',
    change_reason = NULLIF($7, ''),
    updated_by = $6,
    updated_at = NOW(),
    activated_at = NOW()
WHERE scope_type = $1
  AND COALESCE(scope_id::text, '') = $2
  AND role_key = $3
  AND template_kind = $4
  AND locale = $5
  AND version = $8
RETURNING
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

