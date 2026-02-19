-- name: realtime__cleanup_old_events :exec
DELETE FROM realtime_events
WHERE created_at < $1::timestamptz;

