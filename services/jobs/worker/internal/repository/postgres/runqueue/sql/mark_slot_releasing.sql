-- name: runqueue__mark_slot_releasing :exec
UPDATE slots
SET state = 'releasing',
    updated_at = NOW()
WHERE project_id = $1::uuid
  AND lease_owner = $2
  AND state = 'leased';
