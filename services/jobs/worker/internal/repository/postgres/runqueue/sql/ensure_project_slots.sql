-- name: runqueue__ensure_project_slots :exec
INSERT INTO slots (project_id, slot_no, state)
SELECT $1::uuid, gs.slot_no, 'free'
FROM generate_series(1, $2::int) AS gs(slot_no)
ON CONFLICT (project_id, slot_no) DO NOTHING;
