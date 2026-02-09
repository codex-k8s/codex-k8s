-- name: runqueue__mark_slot_free :exec
UPDATE slots
SET state = 'free',
    lease_owner = NULL,
    lease_until = NULL,
    updated_at = NOW()
WHERE project_id = $1::uuid
  AND lease_owner = $2
  AND state IN ('leased', 'releasing');
