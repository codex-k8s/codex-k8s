-- name: changegovernance__upsert_wave :one
INSERT INTO change_governance_waves (
    package_id,
    wave_key,
    publish_order,
    dominant_intent,
    bounded_scope_kind,
    publication_state,
    summary,
    verification_targets_json
)
VALUES (
    $1::uuid,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8::jsonb
)
ON CONFLICT (package_id, wave_key) DO UPDATE
SET
    publish_order = EXCLUDED.publish_order,
    dominant_intent = EXCLUDED.dominant_intent,
    bounded_scope_kind = EXCLUDED.bounded_scope_kind,
    publication_state = EXCLUDED.publication_state,
    summary = EXCLUDED.summary,
    verification_targets_json = EXCLUDED.verification_targets_json,
    updated_at = NOW()
RETURNING
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
    updated_at;
