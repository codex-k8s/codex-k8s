-- name: interactionrequest__count_request_states :many
SELECT
    interaction_kind,
    state,
    COUNT(*)::bigint AS total
FROM interaction_requests
GROUP BY interaction_kind, state;
