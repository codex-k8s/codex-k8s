-- name: staffrun__list_for_user :many
SELECT
    ar.id,
    ar.correlation_id,
    ar.project_id::text AS project_id,
    COALESCE(p.slug, '') AS project_slug,
    COALESCE(p.name, '') AS project_name,
    CASE
        WHEN COALESCE(ar.run_payload->'issue'->>'number', '') ~ '^[0-9]+$'
            THEN (ar.run_payload->'issue'->>'number')::int
        ELSE NULL
    END AS issue_number,
    COALESCE(ar.run_payload->'issue'->>'html_url', '') AS issue_url,
    COALESCE(ar.run_payload->'trigger'->>'kind', '') AS trigger_kind,
    COALESCE(ar.run_payload->'trigger'->>'label', '') AS trigger_label,
    COALESCE(pr.pr_url, '') AS pr_url,
    pr.pr_number,
    ar.status,
    ar.created_at,
    ar.started_at,
    ar.finished_at
FROM agent_runs ar
JOIN project_members pm ON pm.project_id = ar.project_id
JOIN projects p ON p.id = ar.project_id
LEFT JOIN LATERAL (
    SELECT
        COALESCE(fe.payload->>'pr_url', '') AS pr_url,
        CASE
            WHEN COALESCE(fe.payload->>'pr_number', '') ~ '^[0-9]+$'
                THEN (fe.payload->>'pr_number')::int
            ELSE NULL
        END AS pr_number
    FROM flow_events fe
    WHERE fe.correlation_id = ar.correlation_id
      AND fe.event_type IN ('run.pr.created', 'run.pr.updated')
    ORDER BY fe.created_at DESC
    LIMIT 1
) pr ON true
WHERE pm.user_id = $1::uuid
  AND ar.project_id IS NOT NULL
ORDER BY ar.created_at DESC
LIMIT $2;
