-- name: missioncontrol__get_command_by_business_intent :one
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
WHERE project_id = $1
  AND business_intent_key = $2;
