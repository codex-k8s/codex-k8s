-- name: runqueue__claim_running :many
WITH candidates AS (
    SELECT r.id
    FROM agent_runs AS r
    WHERE r.status = 'running'
      AND r.project_id IS NOT NULL
      AND (
            r.lease_owner = $1
         OR r.lease_owner IS NULL
         OR r.lease_until IS NULL
         OR r.lease_until < NOW()
      )
    ORDER BY r.started_at NULLS FIRST, r.created_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT $3
),
claimed AS (
    UPDATE agent_runs AS r
    SET lease_owner = $1,
        lease_until = NOW() + ($2::text)::interval,
        updated_at = NOW()
    FROM candidates AS c
    WHERE r.id = c.id
    RETURNING r.id, r.correlation_id, r.project_id, r.learning_mode, r.run_payload, r.started_at
)
SELECT c.id,
       c.correlation_id,
       c.project_id,
       COALESCE(s.id::text, '') AS slot_id,
       COALESCE(s.slot_no, 0) AS slot_no,
       c.learning_mode,
       c.run_payload,
       c.started_at
FROM claimed AS c
LEFT JOIN slots AS s
  ON s.project_id = c.project_id
 AND s.lease_owner = c.id::text
 AND s.state = 'leased'
ORDER BY c.started_at NULLS FIRST;
