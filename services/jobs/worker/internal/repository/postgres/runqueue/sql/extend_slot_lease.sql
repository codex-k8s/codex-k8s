-- name: runqueue__extend_slot_lease :exec
UPDATE slots
SET lease_until = NOW() + ($3::text)::interval,
    updated_at = NOW()
WHERE project_id = $1::uuid
  AND lease_owner = $2
  AND state = 'leased';
