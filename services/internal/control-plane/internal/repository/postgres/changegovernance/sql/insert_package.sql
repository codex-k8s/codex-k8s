-- name: changegovernance__insert_package :one
INSERT INTO change_governance_packages (
    package_key,
    project_id,
    repository_full_name,
    issue_number,
    pr_number,
    active_projection_version
)
VALUES (
    $1,
    $2::uuid,
    $3,
    $4,
    $5,
    0
)
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
