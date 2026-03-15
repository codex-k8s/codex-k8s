-- name: interactionrequest__list_open_channel_bindings_by_provider_chat_for_update :many
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
JOIN interaction_requests ir
    ON ir.active_channel_binding_id = icb.id
WHERE icb.adapter_kind = 'telegram'
  AND icb.provider_chat_ref = $1
  AND ir.state = 'open'
ORDER BY icb.updated_at DESC
FOR UPDATE;
