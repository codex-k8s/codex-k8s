-- name: agentrun__search_recent_by_project_issue_or_pull_request :many
WITH run_items AS (
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
)
SELECT
    run_id,
    correlation_id,
    project_id,
    repository_full_name,
    agent_key,
    issue_number,
    issue_url,
    pull_request_number,
    pull_request_url,
    trigger_kind,
    trigger_label,
    status,
    created_at,
    started_at,
    finished_at
FROM run_items
WHERE (
        $3::bigint > 0
        AND issue_number = $3::bigint
    )
   OR (
        $4::bigint > 0
        AND pull_request_number = $4::bigint
   )
ORDER BY created_at DESC, run_id DESC
LIMIT $5;
