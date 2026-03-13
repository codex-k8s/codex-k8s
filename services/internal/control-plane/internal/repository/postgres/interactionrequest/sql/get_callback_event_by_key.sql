-- name: interactionrequest__get_callback_event_by_key :one
SELECT
    id,
    interaction_id::text AS interaction_id,
    delivery_id::text AS delivery_id,
    adapter_event_id,
    callback_kind,
    classification,
    normalized_payload_json,
    raw_payload_json,
    received_at,
    processed_at
FROM interaction_callback_events
WHERE interaction_id = $1::uuid
  AND adapter_event_id = $2
LIMIT 1;
