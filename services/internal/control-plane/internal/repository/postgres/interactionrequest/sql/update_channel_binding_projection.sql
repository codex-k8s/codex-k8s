-- name: interactionrequest__update_channel_binding_projection :one
UPDATE interaction_channel_bindings
SET
    continuation_state = $2,
    last_operator_signal_code = $3,
    last_operator_signal_at = $4,
    updated_at = NOW()
WHERE id = $1
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
    updated_at;
