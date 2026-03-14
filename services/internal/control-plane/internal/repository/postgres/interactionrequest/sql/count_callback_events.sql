-- name: interactionrequest__count_callback_events :many
SELECT
    callback_kind,
    classification,
    COUNT(*)::bigint AS total
FROM interaction_callback_events
GROUP BY callback_kind, classification;
