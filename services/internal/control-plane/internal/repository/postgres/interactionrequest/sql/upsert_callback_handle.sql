-- name: interactionrequest__upsert_callback_handle :exec
INSERT INTO interaction_callback_handles (
    interaction_id,
    channel_binding_id,
    handle_hash,
    handle_kind,
    option_id,
    state,
    response_deadline_at,
    grace_expires_at
)
VALUES ($1::uuid, $2, $3, $4, $5, 'open', $6, $7)
ON CONFLICT (handle_hash) DO NOTHING;
