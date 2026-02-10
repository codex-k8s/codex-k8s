-- name: staffrun__delete_events_by_project_id :exec
DELETE FROM flow_events
WHERE correlation_id IN (
    SELECT correlation_id
    FROM agent_runs
    WHERE project_id = $1::uuid
);

