-- name: interactionrequest__count_dispatch_attempts :many
SELECT
    ir.interaction_kind,
    ida.adapter_kind,
    ida.status,
    COUNT(*)::bigint AS total
FROM interaction_delivery_attempts ida
JOIN interaction_requests ir
  ON ir.id = ida.interaction_id
GROUP BY ir.interaction_kind, ida.adapter_kind, ida.status;
