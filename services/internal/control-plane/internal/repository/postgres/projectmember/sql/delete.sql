-- name: projectmember__delete :exec
DELETE FROM project_members
WHERE project_id = $1::uuid
  AND user_id = $2::uuid;

