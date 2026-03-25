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
ORDER BY
    CASE WHEN publication_state = 'superseded' THEN 1 ELSE 0 END ASC,
    CASE
        WHEN publish_order < 0 THEN -publish_order
        ELSE publish_order
    END ASC,
    wave_key ASC;
