-- name: interactionrequest__get_callback_handle_by_hash_for_update :one
SELECT
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
    created_at
FROM interaction_callback_handles
WHERE handle_hash = $1
FOR UPDATE;
