-- name: prompttemplate__list_audit_events :many
SELECT
    id,
    correlation_id,
    COALESCE(payload->>'project_id', '') AS project_id,
    COALESCE(payload->>'template_key', '') AS template_key,
    NULLIF(payload->>'version', '')::integer AS version,
    COALESCE(actor_id, '') AS actor_id,
    event_type,
    payload::text AS payload_json,
    created_at
FROM flow_events
WHERE event_type LIKE 'prompt_template.%'
  AND ($1 = '' OR payload->>'project_id' = $1)
  AND ($2 = '' OR payload->>'template_key' = $2)
  AND ($3 = '' OR actor_id = $3)
ORDER BY created_at DESC
LIMIT $4;

