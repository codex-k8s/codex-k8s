-- name: runqueue__list_running :many
SELECT r.id,
       r.correlation_id,
       r.project_id,
       COALESCE(s.id::text, '')     AS slot_id,
       COALESCE(s.slot_no, 0)      AS slot_no,
       r.learning_mode,
       r.run_payload,
       r.started_at
FROM agent_runs AS r
LEFT JOIN slots AS s
  ON s.project_id = r.project_id
 AND s.lease_owner = r.id::text
 AND s.state = 'leased'
WHERE r.status = 'running'
  AND r.project_id IS NOT NULL
ORDER BY r.started_at NULLS FIRST, r.created_at ASC
LIMIT $1;
