-- name: missioncontrol__update_command_status :one
UPDATE mission_control_commands
SET
    status = $3,
    failure_reason = CASE WHEN $4::boolean THEN $5::text ELSE failure_reason END,
    approval_request_id = CASE WHEN $6::boolean THEN $7::uuid ELSE approval_request_id END,
    approval_state = CASE WHEN $8::boolean THEN $9::text ELSE approval_state END,
    approval_requested_at = CASE WHEN $10::boolean THEN $11::timestamptz ELSE approval_requested_at END,
    approval_decided_at = CASE WHEN $12::boolean THEN $13::timestamptz ELSE approval_decided_at END,
    result_payload = CASE WHEN $14::boolean THEN $15::jsonb ELSE result_payload END,
    provider_delivery_ids = CASE WHEN $16::boolean THEN $17::jsonb ELSE provider_delivery_ids END,
    updated_at = COALESCE($18::timestamptz, NOW()),
    reconciled_at = CASE WHEN $19::boolean THEN $20::timestamptz ELSE reconciled_at END
WHERE project_id = $1
  AND id = $2
RETURNING
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
    reconciled_at;
