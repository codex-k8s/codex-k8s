-- name: projectmember__get_learning_mode_override :one
SELECT learning_mode_override
FROM project_members
WHERE project_id = $1::uuid
  AND user_id = $2::uuid;

