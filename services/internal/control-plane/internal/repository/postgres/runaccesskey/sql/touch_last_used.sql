UPDATE run_access_keys
SET
    last_used_at = $2,
    updated_at = $3
WHERE run_id = $1
  AND status = 'active';
