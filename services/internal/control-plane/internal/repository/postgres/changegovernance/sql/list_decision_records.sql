-- name: changegovernance__list_decision_records :many
SELECT
    id::text AS id,
    package_id::text AS package_id,
    scope_kind,
    scope_ref,
    decision_id,
    decision_kind,
    state,
    actor_kind,
    residual_risk_tier,
    summary_markdown,
    decision_payload_json,
    recorded_at,
    created_at
FROM change_governance_decision_records
WHERE package_id = $1::uuid
ORDER BY recorded_at DESC, id DESC;
