-- name: mcpactionrequest__list_pending :many
SELECT
    mar.id,
    mar.correlation_id,
    mar.run_id::text AS run_id,
    COALESCE(ar.project_id::text, '') AS project_id,
    COALESCE(p.slug, '') AS project_slug,
    COALESCE(p.name, '') AS project_name,
    CASE
        WHEN COALESCE(ar.run_payload->'issue'->>'number', '') ~ '^[0-9]+$'
            THEN (ar.run_payload->'issue'->>'number')::int
        ELSE NULL
    END AS issue_number,
    CASE
        WHEN COALESCE(ar.run_payload->'pull_request'->>'number', '') ~ '^[0-9]+$'
            THEN (ar.run_payload->'pull_request'->>'number')::int
        ELSE NULL
    END AS pr_number,
    COALESCE(ar.run_payload->'trigger'->>'label', '') AS trigger_label,
    mar.tool_name,
    mar.action,
    mar.target_ref,
    mar.approval_mode,
    mar.approval_state,
    mar.requested_by,
    COALESCE(mar.applied_by, '') AS applied_by,
    mar.payload,
    mar.created_at,
    mar.updated_at
FROM mcp_action_requests mar
LEFT JOIN agent_runs ar ON ar.id = mar.run_id
LEFT JOIN projects p ON p.id = ar.project_id
WHERE mar.approval_state = 'requested'
ORDER BY mar.created_at ASC
LIMIT $1;
