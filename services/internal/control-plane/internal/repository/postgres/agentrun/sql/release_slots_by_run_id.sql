-- name: agentrun__release_slots_by_run_id :exec
UPDATE slots
SET state = 'free',
    lease_owner = NULL,
    lease_until = NULL,
    updated_at = NOW()
WHERE lease_owner = $1
  AND state IN ('leased', 'releasing');
