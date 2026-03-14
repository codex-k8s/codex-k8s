-- name: interactionrequest__update_attempt :one
UPDATE interaction_delivery_attempts
SET
    adapter_kind = $2,
    status = $3,
    request_envelope_json = $4::jsonb,
    ack_payload_json = $5::jsonb,
    adapter_delivery_id = $6,
    provider_message_ref_json = $7::jsonb,
    retryable = $8,
    next_retry_at = $9,
    last_error_code = $10,
    finished_at = $11
WHERE delivery_id = $1::uuid
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
