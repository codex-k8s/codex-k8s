-- name: missioncontrol__list_commands_all :many
SELECT
    id::text AS id,
    project_id::text AS project_id,
    command_kind,
    target_entity_id,
    actor_id,
    business_intent_key,
    correlation_id,
    status,
    failure_reason,
    approval_request_id::text AS approval_request_id,
    approval_state,
    approval_requested_at,
    approval_decided_at,
    payload AS payload_json,
    result_payload AS result_payload_json,
    provider_delivery_ids AS provider_deliveries_json,
    requested_at,
    updated_at,
    reconciled_at
FROM mission_control_commands
WHERE (
      COALESCE(array_length($1::text[], 1), 0) = 0
      OR status = ANY($1::text[])
  )
ORDER BY updated_at DESC, requested_at DESC, id DESC
LIMIT $2;
