-- name: interactionrequest__get_callback_event_by_key :one
SELECT
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
    processed_at
FROM interaction_callback_events
WHERE interaction_id = $1::uuid
  AND adapter_event_id = $2
LIMIT 1;
