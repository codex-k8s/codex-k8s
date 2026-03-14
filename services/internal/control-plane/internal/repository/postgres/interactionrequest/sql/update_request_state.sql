-- name: interactionrequest__update_request_state :one
UPDATE interaction_requests
SET
    state = $2,
    resolution_kind = $3,
    effective_response_id = $4,
    updated_at = NOW()
WHERE id = $1::uuid
RETURNING
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
    updated_at;
