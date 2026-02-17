-- name: repocfg__release_preflight_lock :exec
DELETE FROM repository_preflight_locks
WHERE repository_id = $1::uuid
  AND lock_token = $2::uuid;

