-- name: githubratelimitwait__update_wait :one
UPDATE github_rate_limit_waits
SET
    signal_origin = $2,
    operation_class = $3,
    state = $4,
    limit_kind = $5,
    confidence = $6,
    recovery_hint_kind = $7,
    signal_id = $8,
    request_fingerprint = $9,
    correlation_id = $10,
    resume_action_kind = $11,
    resume_payload_json = $12,
    manual_action_kind = $13,
    auto_resume_attempts_used = $14,
    max_auto_resume_attempts = $15,
    resume_not_before = $16,
    last_resume_attempt_at = $17,
    last_signal_at = $18,
    resolved_at = $19,
    updated_at = NOW()
WHERE id = $1::uuid
RETURNING *;
