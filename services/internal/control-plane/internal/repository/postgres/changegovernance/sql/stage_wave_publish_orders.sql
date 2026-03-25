-- name: changegovernance__stage_wave_publish_orders :exec
WITH staged AS (
    SELECT
        id,
        -ROW_NUMBER() OVER (ORDER BY created_at ASC, id ASC) AS staged_publish_order
    FROM change_governance_waves
    WHERE package_id = $1::uuid
)
UPDATE change_governance_waves AS target
SET
    publish_order = staged.staged_publish_order,
    updated_at = NOW()
FROM staged
WHERE target.id = staged.id;
