-- name: changegovernance__list_waves :many
SELECT
    id::text AS id,
    package_id::text AS package_id,
    wave_key,
    publish_order,
    dominant_intent,
    bounded_scope_kind,
    publication_state,
    evidence_completeness_state,
    verification_minimum_state,
    summary,
    verification_targets_json,
    created_at,
    updated_at
FROM change_governance_waves
WHERE package_id = $1::uuid
ORDER BY publish_order ASC, wave_key ASC;
