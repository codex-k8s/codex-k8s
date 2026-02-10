-- name: repocfg__delete :exec
DELETE FROM repositories
WHERE project_id = $1::uuid
  AND id = $2::uuid;

