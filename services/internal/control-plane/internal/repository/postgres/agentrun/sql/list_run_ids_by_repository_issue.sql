-- name: agentrun__list_run_ids_by_repository_issue :many
SELECT ar.id
FROM agent_runs ar
WHERE LOWER(COALESCE(ar.run_payload->'repository'->>'full_name', '')) = LOWER($1)
  AND CASE
        WHEN COALESCE(ar.run_payload->'issue'->>'number', '') ~ '^[0-9]+$'
          THEN (ar.run_payload->'issue'->>'number')::bigint
        ELSE 0
      END = $2::bigint
ORDER BY ar.created_at DESC
LIMIT $3;
