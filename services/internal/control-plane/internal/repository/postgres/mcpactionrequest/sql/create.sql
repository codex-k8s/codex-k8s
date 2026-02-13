-- name: mcpactionrequest__create :one
INSERT INTO mcp_action_requests (
    correlation_id,
    run_id,
    tool_name,
    action,
    target_ref,
    approval_mode,
    approval_state,
    requested_by,
    applied_by,
    payload
)
VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8, $9, $10::jsonb)
RETURNING id;
