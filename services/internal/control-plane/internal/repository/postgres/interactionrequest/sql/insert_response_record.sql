-- name: interactionrequest__insert_response_record :one
INSERT INTO interaction_response_records (
    interaction_id,
    channel_binding_id,
    callback_event_id,
    handle_kind,
    response_kind,
    selected_option_id,
    free_text,
    responder_ref,
    classification,
    is_effective,
    responded_at
)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING
    id,
    interaction_id::text AS interaction_id,
    channel_binding_id,
    callback_event_id,
    handle_kind,
    response_kind,
    COALESCE(selected_option_id, '') AS selected_option_id,
    COALESCE(free_text, '') AS free_text,
    COALESCE(responder_ref, '') AS responder_ref,
    classification,
    is_effective,
    responded_at;
