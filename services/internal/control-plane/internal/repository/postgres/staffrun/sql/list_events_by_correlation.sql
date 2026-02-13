-- name: staffrun__list_events_by_correlation :many
SELECT correlation_id, event_type, created_at, payload::text
FROM flow_events
WHERE correlation_id = $1
ORDER BY created_at DESC
LIMIT $2;
