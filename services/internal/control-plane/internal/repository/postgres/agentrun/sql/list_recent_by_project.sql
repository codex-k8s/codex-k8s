-- name: agentrun__list_recent_by_project :many
SELECT
    ar.id AS run_id,
    ar.correlation_id,
    ar.project_id::text AS project_id,
    COALESCE(ar.run_payload->'repository'->>'full_name', '') AS repository_full_name,
    COALESCE(ar.run_payload->'agent'->>'key', '') AS agent_key,
    CASE
        WHEN COALESCE(ar.run_payload->'issue'->>'number', '') ~ '^[0-9]+$'
            THEN (ar.run_payload->'issue'->>'number')::bigint
        ELSE NULL
    END AS issue_number,
    COALESCE(ar.run_payload->'issue'->>'html_url', '') AS issue_url,
    CASE
        WHEN COALESCE(ar.run_payload->'pull_request'->>'number', '') ~ '^[0-9]+$'
            THEN (ar.run_payload->'pull_request'->>'number')::bigint
        WHEN pr.pr_number > 0
            THEN pr.pr_number::bigint
        ELSE NULL
    END AS pull_request_number,
    CASE
        WHEN COALESCE(ar.run_payload->'pull_request'->>'html_url', '') <> ''
            THEN ar.run_payload->'pull_request'->>'html_url'
        ELSE COALESCE(pr.pr_url, '')
    END AS pull_request_url,
    COALESCE(ar.run_payload->'trigger'->>'kind', '') AS trigger_kind,
    COALESCE(ar.run_payload->'trigger'->>'label', '') AS trigger_label,
    ar.status,
    ar.created_at,
    ar.started_at,
    ar.finished_at
FROM agent_runs ar
LEFT JOIN LATERAL (
    SELECT
        CASE
            WHEN COALESCE(fe.payload->>'pr_number', '') ~ '^[0-9]+$'
                THEN (fe.payload->>'pr_number')::bigint
            ELSE 0
        END AS pr_number,
        COALESCE(fe.payload->>'pr_url', '') AS pr_url
    FROM flow_events fe
    WHERE fe.correlation_id = ar.correlation_id
      AND fe.event_type IN ('run.pr.created', 'run.pr.updated')
    ORDER BY fe.created_at DESC
    LIMIT 1
) pr ON true
WHERE ar.project_id = $1::uuid
  AND (
    NULLIF($2, '') IS NULL
    OR lower(COALESCE(ar.run_payload->'repository'->>'full_name', '')) = lower($2)
  )
ORDER BY ar.created_at DESC, ar.id DESC
LIMIT $3 OFFSET $4;
