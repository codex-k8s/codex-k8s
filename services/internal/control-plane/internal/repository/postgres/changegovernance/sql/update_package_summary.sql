-- name: changegovernance__update_package_summary :one
UPDATE change_governance_packages
SET
    pr_number = $2,
    risk_tier = $3,
    bundle_admissibility = $4,
    publication_state = $5,
    evidence_completeness_state = $6,
    verification_minimum_state = $7,
    waiver_state = $8,
    release_readiness_state = $9,
    governance_feedback_state = $10,
    latest_correlation_id = $11,
    active_projection_version = change_governance_packages.active_projection_version + 1,
    updated_at = COALESCE($12::timestamptz, NOW())
WHERE id = $1::uuid
  AND active_projection_version = $13
RETURNING
    id::text AS id,
    package_key,
    project_id::text AS project_id,
    repository_full_name,
    issue_number,
    pr_number,
    risk_tier,
    bundle_admissibility,
    publication_state,
    evidence_completeness_state,
    verification_minimum_state,
    waiver_state,
    release_readiness_state,
    governance_feedback_state,
    active_projection_version,
    latest_correlation_id,
    created_at,
    updated_at;
