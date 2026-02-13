-- name: mcpactionrequest__update_state :one
UPDATE mcp_action_requests
SET
    approval_state = $2,
    applied_by = $3,
    payload = CASE
        WHEN $4::jsonb = '{}'::jsonb THEN payload
        ELSE payload || $4::jsonb
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING id;
