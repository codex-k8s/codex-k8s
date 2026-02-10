-- name: projectmember__set_learning_mode_override :exec
UPDATE project_members
SET learning_mode_override = $3,
    updated_at = NOW()
WHERE project_id = $1::uuid
  AND user_id = $2::uuid;

