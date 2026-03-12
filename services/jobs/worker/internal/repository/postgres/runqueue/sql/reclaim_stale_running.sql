-- name: runqueue__reclaim_stale_running :many
UPDATE agent_runs AS r
SET lease_owner = $2,
    lease_until = NOW() + ($4::text)::interval,
    updated_at = NOW()
WHERE r.id = $1
  AND r.status = 'running'
  AND r.project_id IS NOT NULL
  AND r.lease_owner = $3
  AND r.lease_until IS NOT NULL
  AND r.lease_until >= NOW()
RETURNING r.id,
          r.correlation_id,
          r.project_id,
          COALESCE((
              SELECT s.id::text
              FROM slots AS s
              WHERE s.project_id = r.project_id
                AND s.lease_owner = r.id::text
                AND s.state = 'leased'
              LIMIT 1
          ), '') AS slot_id,
          COALESCE((
              SELECT s.slot_no
              FROM slots AS s
              WHERE s.project_id = r.project_id
                AND s.lease_owner = r.id::text
                AND s.state = 'leased'
              LIMIT 1
          ), 0) AS slot_no,
          r.learning_mode,
          r.run_payload,
          r.started_at,
          COALESCE(r.lease_owner, '') AS lease_owner,
          r.lease_until;
