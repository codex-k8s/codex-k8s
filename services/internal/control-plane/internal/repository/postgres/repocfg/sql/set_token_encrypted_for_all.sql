-- name: repocfg__set_token_encrypted_for_all :exec
UPDATE repositories
SET token_encrypted = $1,
    updated_at = NOW()
WHERE token_encrypted IS DISTINCT FROM $1;
