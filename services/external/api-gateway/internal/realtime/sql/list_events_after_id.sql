-- name: realtime__list_events_after_id :many
SELECT
    id,
    topic,
    scope,
    payload_json,
    correlation_id,
    COALESCE(project_id::text, '') AS project_id,
    COALESCE(run_id::text, '') AS run_id,
    COALESCE(task_id::text, '') AS task_id,
    created_at
FROM realtime_events
WHERE id > $1::bigint
ORDER BY id ASC
LIMIT $2::int;

