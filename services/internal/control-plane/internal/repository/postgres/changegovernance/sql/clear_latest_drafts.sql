-- name: changegovernance__clear_latest_drafts :exec
UPDATE change_governance_internal_drafts
SET is_latest = false
WHERE package_id = $1::uuid
  AND id <> $2::uuid
  AND is_latest = true;
