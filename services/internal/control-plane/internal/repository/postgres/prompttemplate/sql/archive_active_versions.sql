-- name: prompttemplate__archive_active_versions :exec
UPDATE prompt_templates
SET
    is_active = FALSE,
    status = 'archived',
    updated_by = $6,
    updated_at = NOW()
WHERE scope_type = $1
  AND COALESCE(scope_id::text, '') = $2
  AND role_key = $3
  AND template_kind = $4
  AND locale = $5
  AND status = 'active';

