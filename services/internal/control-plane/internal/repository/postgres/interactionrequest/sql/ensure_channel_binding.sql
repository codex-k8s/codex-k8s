-- name: interactionrequest__ensure_channel_binding :one
WITH existing_binding AS (
    SELECT
        icb.id,
        icb.interaction_id::text AS interaction_id,
        icb.adapter_kind,
        icb.recipient_ref,
        icb.provider_chat_ref,
        icb.provider_message_ref_json,
        icb.callback_token_key_id,
        icb.callback_token_expires_at,
        icb.edit_capability,
        icb.continuation_state,
        icb.last_operator_signal_code,
        icb.last_operator_signal_at,
        icb.created_at,
        icb.updated_at
    FROM interaction_channel_bindings icb
    WHERE icb.interaction_id = $1::uuid
      AND icb.adapter_kind = $2
    LIMIT 1
), inserted_binding AS (
    INSERT INTO interaction_channel_bindings (
        interaction_id,
        adapter_kind,
        recipient_ref,
        callback_token_key_id,
        callback_token_expires_at
    )
    SELECT $1::uuid, $2, $3, $4, $5
    WHERE NOT EXISTS (SELECT 1 FROM existing_binding)
    RETURNING
        id,
        interaction_id::text AS interaction_id,
        adapter_kind,
        recipient_ref,
        provider_chat_ref,
        provider_message_ref_json,
        callback_token_key_id,
        callback_token_expires_at,
        edit_capability,
        continuation_state,
        last_operator_signal_code,
        last_operator_signal_at,
        created_at,
        updated_at
), touched_binding AS (
    UPDATE interaction_channel_bindings
    SET
        recipient_ref = $3,
        callback_token_key_id = COALESCE($4, callback_token_key_id),
        callback_token_expires_at = COALESCE($5, callback_token_expires_at),
        updated_at = NOW()
    WHERE id = COALESCE(
        (SELECT id FROM existing_binding),
        (SELECT id FROM inserted_binding)
    )
    RETURNING
        id,
        interaction_id::text AS interaction_id,
        adapter_kind,
        recipient_ref,
        provider_chat_ref,
        provider_message_ref_json,
        callback_token_key_id,
        callback_token_expires_at,
        edit_capability,
        continuation_state,
        last_operator_signal_code,
        last_operator_signal_at,
        created_at,
        updated_at
), updated_request AS (
    UPDATE interaction_requests
    SET
        channel_family = 'telegram',
        active_channel_binding_id = (SELECT id FROM touched_binding),
        updated_at = NOW()
    WHERE id = $1::uuid
    RETURNING 1
)
SELECT * FROM touched_binding;
