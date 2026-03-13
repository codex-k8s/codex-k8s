-- name: interactionrequest__select_latest_attempt_for_update :one
SELECT
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
    finished_at
FROM interaction_delivery_attempts
WHERE interaction_id = $1::uuid
ORDER BY attempt_no DESC
LIMIT 1
FOR UPDATE;
