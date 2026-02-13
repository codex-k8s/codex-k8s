-- name: staffrun__list_all :many
-- This query builds the runs table for platform admins.
-- It joins run/project metadata and enriches each row with:
-- - newest PR reference;
-- - newest runtime artifacts (job_name, job_namespace, namespace) derived from flow events.
-- LATERAL blocks keep enrichment deterministic by selecting latest non-empty values.
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
    COALESCE(rt.job_name, '') AS job_name,
    COALESCE(rt.job_namespace, '') AS job_namespace,
    COALESCE(rt.namespace, '') AS namespace,
    COALESCE(pr.pr_url, '') AS pr_url,
    pr.pr_number,
    ar.status,
    ar.created_at,
    ar.started_at,
    ar.finished_at
FROM agent_runs ar
LEFT JOIN projects p ON p.id = ar.project_id
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
LEFT JOIN LATERAL (
    SELECT
        COALESCE((
            SELECT COALESCE(fe.payload->>'job_name', '')
            FROM flow_events fe
            WHERE fe.correlation_id = ar.correlation_id
              AND fe.event_type IN ('run.started', 'run.namespace.prepared')
              AND COALESCE(fe.payload->>'job_name', '') <> ''
            ORDER BY fe.created_at DESC
            LIMIT 1
        ), '') AS job_name,
        COALESCE((
            SELECT
                CASE
                    WHEN COALESCE(fe.payload->>'job_namespace', '') <> ''
                        THEN fe.payload->>'job_namespace'
                    WHEN COALESCE(fe.payload->>'namespace', '') <> ''
                        THEN fe.payload->>'namespace'
                    ELSE ''
                END
            FROM flow_events fe
            WHERE fe.correlation_id = ar.correlation_id
              AND fe.event_type IN ('run.started', 'run.namespace.prepared')
              AND (
                    COALESCE(fe.payload->>'job_namespace', '') <> ''
                    OR COALESCE(fe.payload->>'namespace', '') <> ''
              )
            ORDER BY fe.created_at DESC
            LIMIT 1
        ), '') AS job_namespace,
        COALESCE((
            SELECT COALESCE(fe.payload->>'namespace', '')
            FROM flow_events fe
            WHERE fe.correlation_id = ar.correlation_id
              AND fe.event_type IN ('run.started', 'run.namespace.prepared')
              AND COALESCE(fe.payload->>'namespace', '') <> ''
            ORDER BY fe.created_at DESC
            LIMIT 1
        ), '') AS namespace
) rt ON true
ORDER BY created_at DESC
LIMIT $1;
