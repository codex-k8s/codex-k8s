-- name: prompttemplate__lock_latest_version :one
SELECT version
FROM prompt_templates
WHERE scope_type = $1
  -- Для global scope в БД хранится NULL scope_id; снаружи он передается как пустая строка.
  AND COALESCE(scope_id::text, '') = $2
  AND role_key = $3
  AND template_kind = $4
  AND locale = $5
ORDER BY version DESC
LIMIT 1
FOR UPDATE;
