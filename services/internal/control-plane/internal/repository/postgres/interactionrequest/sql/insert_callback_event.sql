-- name: interactionrequest__insert_callback_event :one
INSERT INTO interaction_callback_events (
    interaction_id,
    channel_binding_id,
    delivery_id,
    adapter_event_id,
    callback_kind,
    classification,
    callback_handle_hash,
    normalized_payload_json,
    raw_payload_json,
    provider_message_ref_json,
    provider_update_id,
    provider_callback_query_id,
    received_at,
    processed_at
)
VALUES ($1::uuid, $2, $3::uuid, $4, $5, $6, $7, $8::jsonb, $9::jsonb, $10::jsonb, $11, $12, $13, $14)
RETURNING
    id,
    interaction_id::text AS interaction_id,
    channel_binding_id,
    delivery_id::text AS delivery_id,
    adapter_event_id,
    callback_kind,
    classification,
    callback_handle_hash,
    normalized_payload_json,
    raw_payload_json,
    provider_message_ref_json,
    provider_update_id,
    provider_callback_query_id,
    received_at,
    processed_at;
