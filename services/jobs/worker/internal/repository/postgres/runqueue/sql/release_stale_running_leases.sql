-- name: runqueue__release_stale_running_leases :many
WITH stale_candidates AS (
    SELECT r.id,
           r.correlation_id,
           COALESCE(r.project_id::text, '') AS project_id,
           COALESCE(r.lease_owner, '') AS previous_lease_owner,
           r.lease_until AS previous_lease_until,
           wi.heartbeat_at,
           wi.expires_at,
           CASE
               WHEN wi.worker_id IS NULL THEN 'missing'
               ELSE wi.status
           END AS worker_status
    FROM agent_runs AS r
    LEFT JOIN worker_instances AS wi
      ON wi.worker_id = r.lease_owner
    WHERE r.status = 'running'
      AND r.lease_owner IS NOT NULL
      AND (
            (
                wi.worker_id IS NOT NULL
                AND (
                    wi.status <> 'active'
                    OR wi.expires_at IS NULL
                    OR wi.expires_at < NOW()
                )
            )
         OR (
                wi.worker_id IS NULL
                AND (
                    (
                        $2::boolean
                        AND NOT (r.lease_owner = ANY(COALESCE($3::text[], ARRAY[]::text[])))
                    )
                    OR (
                        NOT $2::boolean
                        AND (
                            r.lease_until IS NULL
                            OR r.lease_until < NOW()
                        )
                    )
                )
            )
      )
    ORDER BY COALESCE(wi.expires_at, r.lease_until, r.updated_at, r.started_at, r.created_at) ASC
    FOR UPDATE OF r SKIP LOCKED
    LIMIT $1
),
released AS (
    UPDATE agent_runs AS r
    SET lease_owner = NULL,
        lease_until = NULL,
        stale_reclaim_pending = TRUE,
        updated_at = NOW()
    FROM stale_candidates AS c
    WHERE r.id = c.id
    RETURNING r.id,
              c.correlation_id,
              c.project_id,
              c.previous_lease_owner,
              c.previous_lease_until,
              c.heartbeat_at,
              c.expires_at,
              c.worker_status
)
SELECT id,
       correlation_id,
       project_id,
       previous_lease_owner,
       previous_lease_until,
       heartbeat_at,
       expires_at,
       worker_status
FROM released
ORDER BY id;
