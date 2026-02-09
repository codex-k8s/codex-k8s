-- name: runqueue__release_expired_slots :exec
UPDATE slots
SET state = 'free',
    lease_owner = NULL,
    lease_until = NULL,
    updated_at = NOW()
WHERE project_id = $1::uuid
  AND state = 'leased'
  AND lease_until IS NOT NULL
  AND lease_until < NOW();
