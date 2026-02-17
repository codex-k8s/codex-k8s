-- name: repocfg__acquire_preflight_lock :one
INSERT INTO repository_preflight_locks (
    repository_id,
    lock_token,
    locked_by_user_id,
    locked_until
)
VALUES (
    $1::uuid,
    $2::uuid,
    NULLIF($3::text, '')::uuid,
    $4::timestamptz
)
ON CONFLICT (repository_id) DO UPDATE
SET lock_token = EXCLUDED.lock_token,
    locked_by_user_id = EXCLUDED.locked_by_user_id,
    locked_until = EXCLUDED.locked_until,
    updated_at = NOW()
WHERE repository_preflight_locks.locked_until < NOW()
RETURNING lock_token::text;

