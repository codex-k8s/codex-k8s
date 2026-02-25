-- name: prompttemplate__get_version :one
SELECT
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
  AND COALESCE(scope_id::text, '') = $2
  AND role_key = $3
  AND template_kind = $4
  AND locale = $5
  AND version = $6
LIMIT 1;

