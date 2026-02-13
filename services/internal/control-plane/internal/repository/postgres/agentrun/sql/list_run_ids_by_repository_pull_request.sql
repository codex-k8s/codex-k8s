-- name: agentrun__list_run_ids_by_repository_pull_request :many
SELECT ar.id
FROM agent_runs ar
JOIN LATERAL (
    SELECT
        CASE
            WHEN COALESCE(fe.payload->>'pr_number', '') ~ '^[0-9]+$'
                THEN (fe.payload->>'pr_number')::bigint
            ELSE NULL
        END AS pr_number
    FROM flow_events fe
    WHERE fe.correlation_id = ar.correlation_id
      AND fe.event_type IN ('run.pr.created', 'run.pr.updated')
    ORDER BY fe.created_at DESC
    LIMIT 1
) pr ON true
WHERE LOWER(COALESCE(ar.run_payload->'repository'->>'full_name', '')) = LOWER($1)
  AND pr.pr_number = $2::bigint
ORDER BY ar.created_at DESC
LIMIT $3;
