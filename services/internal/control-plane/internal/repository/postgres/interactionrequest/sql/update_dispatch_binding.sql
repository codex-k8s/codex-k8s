-- name: interactionrequest__update_dispatch_binding :one
UPDATE interaction_channel_bindings
SET
    provider_chat_ref = COALESCE(NULLIF($2::jsonb->>'chat_ref', ''), provider_chat_ref),
    provider_message_ref_json = $2::jsonb,
    edit_capability = $3,
    callback_token_expires_at = COALESCE($4, callback_token_expires_at),
    continuation_state = CASE
        WHEN $3 = 'follow_up_only' THEN 'follow_up_required'
        ELSE 'ready_for_edit'
    END,
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
