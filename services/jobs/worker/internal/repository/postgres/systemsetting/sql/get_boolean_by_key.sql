-- name: workersystemsetting__get_boolean_by_key :one
SELECT value_json
FROM system_settings
WHERE key = $1
  AND value_kind = 'boolean';
