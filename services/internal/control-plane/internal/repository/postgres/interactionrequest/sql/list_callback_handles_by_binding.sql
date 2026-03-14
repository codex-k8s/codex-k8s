-- name: interactionrequest__list_callback_handles_by_binding :many
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
WHERE channel_binding_id = $1
ORDER BY id;
