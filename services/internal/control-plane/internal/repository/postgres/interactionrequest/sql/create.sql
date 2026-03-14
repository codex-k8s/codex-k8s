-- name: interactionrequest__create :one
INSERT INTO interaction_requests (
    project_id,
    run_id,
    interaction_kind,
    channel_family,
    state,
    resolution_kind,
    recipient_provider,
    recipient_ref,
    request_payload_json,
    context_links_json,
    response_deadline_at
)
VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb, $11)
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
