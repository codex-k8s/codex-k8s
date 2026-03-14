-- name: interactionrequest__mark_callback_handle_used :one
UPDATE interaction_callback_handles
SET
    state = 'used',
    used_callback_event_id = $2,
    used_at = $3
WHERE id = $1
RETURNING
    id,
    interaction_id::text AS interaction_id,
    channel_binding_id,
    handle_hash,
    handle_kind,
    option_id,
    state,
    response_deadline_at,
    grace_expires_at,
    used_callback_event_id,
    used_at,
    created_at;
