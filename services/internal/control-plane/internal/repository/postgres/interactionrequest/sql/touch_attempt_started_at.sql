-- name: interactionrequest__touch_attempt_started_at :one
UPDATE interaction_delivery_attempts
SET
    started_at = $2,
    finished_at = NULL
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
