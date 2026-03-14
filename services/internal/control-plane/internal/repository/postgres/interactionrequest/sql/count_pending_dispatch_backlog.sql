-- name: interactionrequest__count_pending_dispatch_backlog :many
SELECT
    interaction_kind,
    CASE
        WHEN last_delivery_attempt_no > 0 THEN 'retry'
        ELSE 'initial'
    END AS queue_kind,
    COUNT(*)::bigint AS total
FROM interaction_requests
WHERE state = 'pending_dispatch'
GROUP BY interaction_kind, queue_kind;
