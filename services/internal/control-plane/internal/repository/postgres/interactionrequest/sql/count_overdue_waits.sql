-- name: interactionrequest__count_overdue_waits :many
SELECT
    interaction_kind,
    COUNT(*)::bigint AS total
FROM interaction_requests
WHERE state IN ('pending_dispatch', 'open')
  AND response_deadline_at IS NOT NULL
  AND response_deadline_at < $1::timestamptz
GROUP BY interaction_kind;
