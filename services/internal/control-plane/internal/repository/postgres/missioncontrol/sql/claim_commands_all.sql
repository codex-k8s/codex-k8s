-- name: missioncontrol__claim_commands_all :many
WITH candidates AS (
    SELECT id
    FROM mission_control_commands
    WHERE (
            COALESCE(array_length($1::text[], 1), 0) = 0
            OR status = ANY($1::text[])
        )
      AND (lease_until IS NULL OR lease_until <= NOW())
    ORDER BY updated_at DESC, requested_at DESC, id DESC
    FOR UPDATE SKIP LOCKED
    LIMIT $2
),
claimed AS (
    UPDATE mission_control_commands AS cmd
    SET
        lease_owner = NULLIF(BTRIM($3::text), ''),
        lease_until = NOW() + make_interval(secs => GREATEST($4::integer, 1))
    FROM candidates
    WHERE cmd.id = candidates.id
    RETURNING
        cmd.id::text AS id,
        cmd.project_id::text AS project_id,
        cmd.command_kind,
        cmd.target_entity_id,
        cmd.actor_id,
        cmd.business_intent_key,
        cmd.correlation_id,
        cmd.status,
        cmd.failure_reason,
        cmd.approval_request_id::text AS approval_request_id,
        cmd.approval_state,
        cmd.approval_requested_at,
        cmd.approval_decided_at,
        cmd.payload AS payload_json,
        cmd.result_payload AS result_payload_json,
        cmd.provider_delivery_ids AS provider_deliveries_json,
        cmd.lease_owner,
        cmd.lease_until,
        cmd.requested_at,
        cmd.updated_at,
        cmd.reconciled_at
)
SELECT *
FROM claimed
ORDER BY updated_at DESC, requested_at DESC, id DESC;
