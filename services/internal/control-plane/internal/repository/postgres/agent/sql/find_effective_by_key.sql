-- name: agent__find_effective_by_key :one
-- Prefer project-scoped active agent config over global fallback for the same key.
SELECT
    id,
    agent_key,
    role_kind,
    project_id::text,
    name
FROM agents
WHERE is_active = TRUE
  AND agent_key = $1
  AND (
    project_id IS NULL
    OR project_id = NULLIF($2, '')::uuid
  )
ORDER BY
    CASE
        WHEN project_id IS NULL THEN 1
        ELSE 0
    END,
    updated_at DESC
LIMIT 1;
