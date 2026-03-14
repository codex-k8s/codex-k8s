-- name: interactionrequest__claim_next_expiry_candidate :one
SELECT
    id,
    project_id::text AS project_id,
    run_id::text AS run_id,
    interaction_kind,
    channel_family,
    state,
    resolution_kind,
    recipient_provider,
    recipient_ref,
    request_payload_json,
    context_links_json,
    response_deadline_at,
    effective_response_id,
    active_channel_binding_id,
    operator_state,
    operator_signal_code,
    operator_signal_at,
    last_delivery_attempt_no,
    created_at,
    updated_at
FROM interaction_requests
WHERE interaction_kind = 'decision_request'
  AND state IN ('pending_dispatch', 'open')
  AND response_deadline_at IS NOT NULL
  AND response_deadline_at <= $1
ORDER BY response_deadline_at, created_at, id
LIMIT 1
FOR UPDATE SKIP LOCKED;
