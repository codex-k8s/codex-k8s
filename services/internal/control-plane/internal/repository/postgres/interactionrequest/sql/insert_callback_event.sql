-- name: interactionrequest__insert_callback_event :one
INSERT INTO interaction_callback_events (
    interaction_id,
    delivery_id,
    adapter_event_id,
    callback_kind,
    classification,
    normalized_payload_json,
    raw_payload_json,
    received_at,
    processed_at
)
VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6::jsonb, $7::jsonb, $8, $9)
RETURNING
    id,
    interaction_id::text AS interaction_id,
    delivery_id::text AS delivery_id,
    adapter_event_id,
    callback_kind,
    classification,
    normalized_payload_json,
    raw_payload_json,
    received_at,
    processed_at;
