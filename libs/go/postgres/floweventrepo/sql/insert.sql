-- name: flowevent__insert :exec
INSERT INTO flow_events (
    correlation_id,
    actor_type,
    actor_id,
    event_type,
    payload,
    created_at
)
VALUES (
    $1,
    $2,
    NULLIF($3, ''),
    $4,
    $5::jsonb,
    $6
);

