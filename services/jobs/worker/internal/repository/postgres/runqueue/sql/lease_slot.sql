-- name: runqueue__lease_slot :one
-- Atomically lease one free slot using SKIP LOCKED to avoid worker contention.
WITH candidate AS (
    SELECT id
    FROM slots
    WHERE project_id = $1::uuid
      AND state = 'free'
    ORDER BY slot_no ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
UPDATE slots
SET state = 'leased',
    lease_owner = $2,
    lease_until = NOW() + ($3::text)::interval,
    updated_at = NOW()
WHERE id = (SELECT id FROM candidate)
RETURNING id, slot_no;
