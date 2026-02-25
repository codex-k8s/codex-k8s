-- name: prompttemplate__list_keys :many
SELECT
    CASE
        WHEN scope_type = 'project'
            THEN 'project/' || scope_id::text || '/' || role_key || '/' || template_kind || '/' || locale
        ELSE
            'global/' || role_key || '/' || template_kind || '/' || locale
    END AS template_key,
    scope_type,
    scope_id::text,
    role_key,
    template_kind,
    locale,
    COALESCE(MAX(version) FILTER (WHERE status = 'active'), 0) AS active_version,
    MAX(updated_at) AS updated_at
FROM prompt_templates
WHERE ($1 = '' OR scope_type = $1)
  AND ($2 = '' OR COALESCE(scope_id::text, '') = $2)
  AND ($3 = '' OR role_key = $3)
  AND ($4 = '' OR template_kind = $4)
  AND ($5 = '' OR locale = $5)
GROUP BY scope_type, scope_id, role_key, template_kind, locale
ORDER BY MAX(updated_at) DESC
LIMIT $6;

