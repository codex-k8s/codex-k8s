-- name: prompttemplate__get_active_version :one
SELECT
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
    activated_at
FROM prompt_templates
WHERE scope_type = $1
  -- For global scope, the DB stores NULL scope_id; externally it is passed as an empty string.
  AND COALESCE(scope_id::text, '') = $2
  AND role_key = $3
  AND template_kind = $4
  AND locale = $5
  AND status = 'active'
ORDER BY version DESC
LIMIT 1;
