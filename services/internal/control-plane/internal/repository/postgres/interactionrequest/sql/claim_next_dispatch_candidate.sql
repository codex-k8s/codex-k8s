-- name: interactionrequest__claim_next_dispatch_candidate :one
SELECT
    ir.id,
    ir.project_id::text AS project_id,
    ir.run_id::text AS run_id,
    ir.interaction_kind,
    ir.channel_family,
    ir.state,
    ir.resolution_kind,
    ir.recipient_provider,
    ir.recipient_ref,
    ir.request_payload_json,
    ir.context_links_json,
    ir.response_deadline_at,
    ir.effective_response_id,
    ir.active_channel_binding_id,
    ir.operator_state,
    ir.operator_signal_code,
    ir.operator_signal_at,
    ir.last_delivery_attempt_no,
    ir.created_at,
    ir.updated_at
FROM interaction_requests ir
LEFT JOIN LATERAL (
    SELECT
        ida.id,
        ida.status,
        ida.next_retry_at,
        ida.started_at
    FROM interaction_delivery_attempts ida
    WHERE ida.interaction_id = ir.id
    ORDER BY ida.attempt_no DESC
    LIMIT 1
) latest_attempt ON true
WHERE ir.state = 'pending_dispatch'
  AND (
      latest_attempt.id IS NULL
      OR (
          latest_attempt.status = 'failed'
          AND latest_attempt.next_retry_at IS NOT NULL
          AND latest_attempt.next_retry_at <= $1
      )
      OR (
          latest_attempt.status = 'pending'
          AND latest_attempt.started_at <= $2
      )
  )
ORDER BY COALESCE(latest_attempt.next_retry_at, ir.created_at), ir.created_at, ir.id
LIMIT 1
FOR UPDATE OF ir SKIP LOCKED;
