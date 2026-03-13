-- name: interactionrequest__find_open_decision_by_run_id :one
SELECT
    id,
    project_id::text AS project_id,
    run_id::text AS run_id,
    interaction_kind,
    state,
    resolution_kind,
    recipient_provider,
    recipient_ref,
    request_payload_json,
    context_links_json,
    response_deadline_at,
    effective_response_id,
    last_delivery_attempt_no,
    created_at,
    updated_at
FROM interaction_requests
WHERE run_id = $1::uuid
  AND interaction_kind = 'decision_request'
  AND state IN ('pending_dispatch', 'open')
ORDER BY created_at DESC
LIMIT 1;
