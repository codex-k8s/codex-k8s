-- name: changegovernance__get_package_by_key_for_update :one
SELECT
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
    updated_at
FROM change_governance_packages
WHERE project_id = $1::uuid
  AND package_key = $2
FOR UPDATE;
