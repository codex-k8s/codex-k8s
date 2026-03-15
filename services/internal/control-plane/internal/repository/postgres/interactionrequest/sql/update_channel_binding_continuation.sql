-- name: interactionrequest__update_channel_binding_continuation :one
UPDATE interaction_channel_bindings
SET
    provider_chat_ref = CASE
        WHEN $2::jsonb = '{}'::jsonb THEN provider_chat_ref
        ELSE COALESCE(NULLIF($2::jsonb->>'chat_ref', ''), provider_chat_ref)
    END,
    provider_message_ref_json = CASE
        WHEN $2::jsonb = '{}'::jsonb THEN provider_message_ref_json
        ELSE $2::jsonb
    END,
    continuation_state = $3,
    last_operator_signal_code = $4,
    last_operator_signal_at = $5,
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
