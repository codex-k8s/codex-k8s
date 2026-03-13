-- name: interactionrequest__create_delivery_attempt :one
INSERT INTO interaction_delivery_attempts (
    interaction_id,
    attempt_no,
    adapter_kind,
    status,
    request_envelope_json,
    ack_payload_json,
    adapter_delivery_id,
    retryable,
    next_retry_at,
    last_error_code,
    started_at,
    finished_at
)
VALUES ($1::uuid, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $9, $10, $11, $12)
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
