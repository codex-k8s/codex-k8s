-- name: interactionrequest__create_delivery_attempt :one
INSERT INTO interaction_delivery_attempts (
    interaction_id,
    channel_binding_id,
    attempt_no,
    adapter_kind,
    delivery_role,
    status,
    request_envelope_json,
    ack_payload_json,
    adapter_delivery_id,
    provider_message_ref_json,
    retryable,
    next_retry_at,
    last_error_code,
    continuation_reason,
    started_at,
    finished_at
)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9, $10::jsonb, $11, $12, $13, $14, $15, $16)
RETURNING
    id,
    interaction_id::text AS interaction_id,
    channel_binding_id,
    attempt_no,
    delivery_id::text AS delivery_id,
    adapter_kind,
    delivery_role,
    status,
    request_envelope_json,
    COALESCE(ack_payload_json, '{}'::jsonb) AS ack_payload_json,
    adapter_delivery_id,
    provider_message_ref_json,
    retryable,
    next_retry_at,
    last_error_code,
    continuation_reason,
    started_at,
    finished_at;
