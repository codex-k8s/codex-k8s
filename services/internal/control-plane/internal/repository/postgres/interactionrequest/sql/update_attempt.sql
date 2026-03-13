-- name: interactionrequest__update_attempt :one
UPDATE interaction_delivery_attempts
SET
    adapter_kind = $2,
    status = $3,
    request_envelope_json = $4::jsonb,
    ack_payload_json = $5::jsonb,
    adapter_delivery_id = $6,
    retryable = $7,
    next_retry_at = $8,
    last_error_code = $9,
    finished_at = $10
WHERE delivery_id = $1::uuid
RETURNING
    id,
    interaction_id::text AS interaction_id,
    attempt_no,
    delivery_id::text AS delivery_id,
    adapter_kind,
    status,
    request_envelope_json,
    COALESCE(ack_payload_json, '{}'::jsonb) AS ack_payload_json,
    adapter_delivery_id,
    retryable,
    next_retry_at,
    last_error_code,
    started_at,
    finished_at;
