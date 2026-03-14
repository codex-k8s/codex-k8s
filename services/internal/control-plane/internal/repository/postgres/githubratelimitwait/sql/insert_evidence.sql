-- name: githubratelimitwait__insert_evidence :one
INSERT INTO github_rate_limit_wait_evidence (
    wait_id,
    event_kind,
    signal_id,
    signal_origin,
    provider_status_code,
    retry_after_seconds,
    rate_limit_limit,
    rate_limit_remaining,
    rate_limit_used,
    rate_limit_reset_at,
    rate_limit_resource,
    github_request_id,
    documentation_url,
    message_excerpt,
    stderr_excerpt,
    payload_json,
    observed_at
)
VALUES (
    $1::uuid,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17
)
RETURNING *;
