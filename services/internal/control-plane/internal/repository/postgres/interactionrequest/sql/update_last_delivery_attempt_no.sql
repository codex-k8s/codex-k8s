-- name: interactionrequest__update_last_delivery_attempt_no :exec
UPDATE interaction_requests
SET
    last_delivery_attempt_no = $2,
    updated_at = NOW()
WHERE id = $1::uuid;
