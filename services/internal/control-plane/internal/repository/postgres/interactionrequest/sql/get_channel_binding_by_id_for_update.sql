-- name: interactionrequest__get_channel_binding_by_id_for_update :one
SELECT
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
FROM interaction_channel_bindings
WHERE id = $1
FOR UPDATE;
