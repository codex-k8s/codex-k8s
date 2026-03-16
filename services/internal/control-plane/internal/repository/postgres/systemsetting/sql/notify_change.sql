-- name: systemsetting__notify_change :exec
SELECT pg_notify('codex_system_settings', json_build_object('key', $1, 'version', $2)::text);
