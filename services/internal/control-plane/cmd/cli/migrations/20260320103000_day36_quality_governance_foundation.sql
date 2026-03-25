-- +goose Up

CREATE TABLE IF NOT EXISTS change_governance_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_key TEXT NOT NULL,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    repository_full_name TEXT NOT NULL,
    issue_number INTEGER NOT NULL,
    pr_number INTEGER NULL,
    risk_tier TEXT NULL,
    bundle_admissibility TEXT NOT NULL DEFAULT 'requires_decomposition',
    publication_state TEXT NOT NULL DEFAULT 'hidden_draft',
    evidence_completeness_state TEXT NOT NULL DEFAULT 'not_started',
    verification_minimum_state TEXT NOT NULL DEFAULT 'not_started',
    waiver_state TEXT NOT NULL DEFAULT 'none',
    release_readiness_state TEXT NOT NULL DEFAULT 'not_ready',
    governance_feedback_state TEXT NOT NULL DEFAULT 'none',
    active_projection_version BIGINT NOT NULL DEFAULT 0,
    latest_correlation_id TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_change_governance_packages_project_package_key UNIQUE (project_id, package_key),
    CONSTRAINT chk_change_governance_packages_risk_tier
        CHECK (risk_tier IS NULL OR risk_tier IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_change_governance_packages_bundle_admissibility
        CHECK (bundle_admissibility IN ('single_wave', 'mechanical_bounded_scope', 'requires_decomposition')),
    CONSTRAINT chk_change_governance_packages_publication_state
        CHECK (publication_state IN ('hidden_draft', 'wave_map_defined', 'waves_published', 'review_ready', 'release_decided', 'feedback_open', 'closed')),
    CONSTRAINT chk_change_governance_packages_evidence_completeness_state
        CHECK (evidence_completeness_state IN ('not_started', 'partial', 'complete', 'gapped', 'waived')),
    CONSTRAINT chk_change_governance_packages_verification_minimum_state
        CHECK (verification_minimum_state IN ('not_started', 'in_progress', 'met', 'failed', 'waived')),
    CONSTRAINT chk_change_governance_packages_waiver_state
        CHECK (waiver_state IN ('none', 'requested', 'approved', 'rejected', 'expired')),
    CONSTRAINT chk_change_governance_packages_release_readiness_state
        CHECK (release_readiness_state IN ('not_ready', 'conditionally_ready', 'ready', 'blocked', 'released')),
    CONSTRAINT chk_change_governance_packages_governance_feedback_state
        CHECK (governance_feedback_state IN ('none', 'open', 'reclassified', 'closed')),
    CONSTRAINT chk_change_governance_packages_projection_version
        CHECK (active_projection_version >= 0),
    CONSTRAINT chk_change_governance_packages_high_risk_requires_waiver
        CHECK (
            risk_tier IS NULL
            OR risk_tier NOT IN ('high', 'critical')
            OR waiver_state = 'approved'
            OR release_readiness_state NOT IN ('conditionally_ready', 'ready', 'released')
        )
);

CREATE INDEX IF NOT EXISTS idx_change_governance_packages_repository_issue
    ON change_governance_packages (repository_full_name, issue_number);

CREATE INDEX IF NOT EXISTS idx_change_governance_packages_pr_number
    ON change_governance_packages (pr_number)
    WHERE pr_number IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_change_governance_packages_latest_correlation
    ON change_governance_packages (latest_correlation_id)
    WHERE latest_correlation_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_change_governance_packages_queue
    ON change_governance_packages (project_id, risk_tier, publication_state, updated_at DESC, id);

CREATE INDEX IF NOT EXISTS idx_change_governance_packages_ready_queue
    ON change_governance_packages (project_id, updated_at DESC, id)
    WHERE publication_state IN ('waves_published', 'review_ready');

CREATE INDEX IF NOT EXISTS idx_change_governance_packages_high_critical_blockers
    ON change_governance_packages (project_id, updated_at DESC, id)
    WHERE risk_tier IN ('high', 'critical')
      AND release_readiness_state IN ('not_ready', 'blocked', 'conditionally_ready');

CREATE TABLE IF NOT EXISTS change_governance_internal_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES change_governance_packages(id) ON DELETE CASCADE,
    run_id UUID NULL REFERENCES agent_runs(id) ON DELETE SET NULL,
    signal_id TEXT NOT NULL UNIQUE,
    draft_ref TEXT NOT NULL,
    draft_checksum TEXT NULL,
    draft_kind TEXT NOT NULL DEFAULT 'internal_working_draft',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_latest BOOLEAN NOT NULL DEFAULT true,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_change_governance_internal_drafts_kind
        CHECK (draft_kind = 'internal_working_draft')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_change_governance_internal_drafts_latest
    ON change_governance_internal_drafts (package_id)
    WHERE is_latest = true;

CREATE INDEX IF NOT EXISTS idx_change_governance_internal_drafts_package_occurred
    ON change_governance_internal_drafts (package_id, occurred_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS change_governance_waves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES change_governance_packages(id) ON DELETE CASCADE,
    wave_key TEXT NOT NULL,
    publish_order INTEGER NOT NULL,
    dominant_intent TEXT NOT NULL,
    bounded_scope_kind TEXT NOT NULL,
    publication_state TEXT NOT NULL DEFAULT 'planned',
    evidence_completeness_state TEXT NOT NULL DEFAULT 'not_started',
    verification_minimum_state TEXT NOT NULL DEFAULT 'not_started',
    summary TEXT NOT NULL,
    verification_targets_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_change_governance_waves_package_wave_key UNIQUE (package_id, wave_key),
    CONSTRAINT uq_change_governance_waves_package_publish_order UNIQUE (package_id, publish_order),
    CONSTRAINT chk_change_governance_waves_dominant_intent
        CHECK (dominant_intent IN ('code_behavior', 'schema', 'transport', 'ui', 'ops', 'mechanical_refactor', 'docs_only')),
    CONSTRAINT chk_change_governance_waves_bounded_scope_kind
        CHECK (bounded_scope_kind IN ('single_context', 'cross_context', 'mechanical_bounded_scope')),
    CONSTRAINT chk_change_governance_waves_publication_state
        CHECK (publication_state IN ('planned', 'published', 'reviewed', 'merged', 'superseded')),
    CONSTRAINT chk_change_governance_waves_evidence_completeness_state
        CHECK (evidence_completeness_state IN ('not_started', 'partial', 'complete', 'gapped', 'waived')),
    CONSTRAINT chk_change_governance_waves_verification_minimum_state
        CHECK (verification_minimum_state IN ('not_started', 'in_progress', 'met', 'failed', 'waived')),
    CONSTRAINT chk_change_governance_waves_mechanical_scope
        CHECK (
            bounded_scope_kind <> 'mechanical_bounded_scope'
            OR dominant_intent IN ('mechanical_refactor', 'docs_only')
        )
);

CREATE INDEX IF NOT EXISTS idx_change_governance_waves_package_order
    ON change_governance_waves (package_id, publish_order, id);

CREATE TABLE IF NOT EXISTS change_governance_evidence_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES change_governance_packages(id) ON DELETE CASCADE,
    wave_id UUID NULL REFERENCES change_governance_waves(id) ON DELETE CASCADE,
    block_kind TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'missing',
    verification_state TEXT NOT NULL DEFAULT 'not_started',
    required_by_tier BOOLEAN NOT NULL DEFAULT false,
    source_kind TEXT NOT NULL,
    artifact_links_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    latest_signal_id TEXT NULL,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_change_governance_evidence_blocks_kind
        CHECK (block_kind IN ('intent_contract', 'verification', 'review_waiver', 'release_readiness', 'runtime_feedback')),
    CONSTRAINT chk_change_governance_evidence_blocks_state
        CHECK (state IN ('missing', 'present', 'verified', 'waived', 'stale')),
    CONSTRAINT chk_change_governance_evidence_blocks_verification_state
        CHECK (verification_state IN ('not_started', 'in_progress', 'met', 'failed', 'waived')),
    CONSTRAINT chk_change_governance_evidence_blocks_source_kind
        CHECK (source_kind IN ('agent_signal', 'github_webhook', 'staff_command', 'worker_feedback', 'backfill_inferred'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_change_governance_evidence_blocks_scope_kind
    ON change_governance_evidence_blocks (
        package_id,
        COALESCE(wave_id, '00000000-0000-0000-0000-000000000000'::uuid),
        block_kind
    );

CREATE INDEX IF NOT EXISTS idx_change_governance_evidence_blocks_package_observed
    ON change_governance_evidence_blocks (package_id, observed_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS change_governance_decision_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES change_governance_packages(id) ON DELETE CASCADE,
    scope_kind TEXT NOT NULL DEFAULT 'package',
    scope_ref TEXT NOT NULL,
    decision_id TEXT NOT NULL UNIQUE,
    decision_kind TEXT NOT NULL,
    state TEXT NOT NULL,
    actor_kind TEXT NOT NULL,
    residual_risk_tier TEXT NULL,
    summary_markdown TEXT NOT NULL,
    decision_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_change_governance_decision_records_scope_kind
        CHECK (scope_kind IN ('package', 'wave', 'evidence_block')),
    CONSTRAINT chk_change_governance_decision_records_kind
        CHECK (decision_kind IN ('risk_classification', 'reclassification', 'waiver', 'release_readiness')),
    CONSTRAINT chk_change_governance_decision_records_state
        CHECK (state IN ('proposed', 'approved', 'rejected', 'superseded')),
    CONSTRAINT chk_change_governance_decision_records_actor_kind
        CHECK (actor_kind IN ('owner', 'reviewer', 'operator', 'system')),
    CONSTRAINT chk_change_governance_decision_records_residual_risk_tier
        CHECK (residual_risk_tier IS NULL OR residual_risk_tier IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT chk_change_governance_decision_records_high_risk_waiver
        CHECK (
            decision_kind <> 'waiver'
            OR state <> 'approved'
            OR residual_risk_tier IS NOT NULL
        )
);

CREATE INDEX IF NOT EXISTS idx_change_governance_decision_records_package_recorded
    ON change_governance_decision_records (package_id, recorded_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS change_governance_feedback_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES change_governance_packages(id) ON DELETE CASCADE,
    feedback_id TEXT NOT NULL UNIQUE,
    gap_kind TEXT NOT NULL,
    source_kind TEXT NOT NULL,
    severity TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'open',
    suggested_action TEXT NOT NULL,
    summary_markdown TEXT NOT NULL,
    related_artifact_ref TEXT NULL,
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_change_governance_feedback_records_gap_kind
        CHECK (gap_kind IN ('under_classified', 'missing_evidence', 'verification_bypass', 'silent_waiver_attempt', 'semantic_mix', 'late_reclassification')),
    CONSTRAINT chk_change_governance_feedback_records_source_kind
        CHECK (source_kind IN ('review', 'release', 'postdeploy', 'remediation', 'worker_sweep', 'backfill')),
    CONSTRAINT chk_change_governance_feedback_records_severity
        CHECK (severity IN ('medium', 'high', 'critical')),
    CONSTRAINT chk_change_governance_feedback_records_state
        CHECK (state IN ('open', 'acknowledged', 'reclassified', 'closed')),
    CONSTRAINT chk_change_governance_feedback_records_suggested_action
        CHECK (suggested_action IN ('reclassify', 'request_evidence', 'record_waiver', 'block_release', 'close_gap')),
    CONSTRAINT chk_change_governance_feedback_records_closed_at
        CHECK ((state = 'closed' AND closed_at IS NOT NULL) OR (state <> 'closed'))
);

CREATE INDEX IF NOT EXISTS idx_change_governance_feedback_records_gap_queue
    ON change_governance_feedback_records (state, severity, opened_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_change_governance_feedback_records_package_opened
    ON change_governance_feedback_records (package_id, opened_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS change_governance_projection_snapshots (
    id BIGSERIAL PRIMARY KEY,
    package_id UUID NOT NULL REFERENCES change_governance_packages(id) ON DELETE CASCADE,
    projection_kind TEXT NOT NULL,
    projection_version BIGINT NOT NULL,
    is_current BOOLEAN NOT NULL DEFAULT true,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    refreshed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_change_governance_projection_snapshots_version
        UNIQUE (package_id, projection_kind, projection_version),
    CONSTRAINT chk_change_governance_projection_snapshots_kind
        CHECK (projection_kind IN ('package_list', 'package_detail', 'operator_gap_queue', 'release_gate', 'github_status_comment'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_change_governance_projection_snapshots_current
    ON change_governance_projection_snapshots (package_id, projection_kind)
    WHERE is_current = true;

CREATE INDEX IF NOT EXISTS idx_change_governance_projection_snapshots_package_kind
    ON change_governance_projection_snapshots (package_id, projection_kind, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS change_governance_artifact_links (
    id BIGSERIAL PRIMARY KEY,
    package_id UUID NOT NULL REFERENCES change_governance_packages(id) ON DELETE CASCADE,
    artifact_kind TEXT NOT NULL,
    artifact_ref TEXT NOT NULL,
    relation_kind TEXT NOT NULL,
    display_label TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_change_governance_artifact_links_artifact_kind
        CHECK (artifact_kind IN ('issue', 'pull_request', 'run', 'agent_session', 'document', 'service_comment', 'release_note')),
    CONSTRAINT chk_change_governance_artifact_links_relation_kind
        CHECK (relation_kind IN ('primary_context', 'evidence_source', 'decision_followup', 'feedback_source'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_change_governance_artifact_links_exact
    ON change_governance_artifact_links (package_id, artifact_kind, artifact_ref, relation_kind);

CREATE UNIQUE INDEX IF NOT EXISTS uq_change_governance_artifact_links_primary_issue
    ON change_governance_artifact_links (package_id)
    WHERE artifact_kind = 'issue'
      AND relation_kind = 'primary_context';

CREATE INDEX IF NOT EXISTS idx_change_governance_artifact_links_package_created
    ON change_governance_artifact_links (package_id, created_at DESC, id DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_change_governance_artifact_links_package_created;
DROP INDEX IF EXISTS uq_change_governance_artifact_links_primary_issue;
DROP INDEX IF EXISTS uq_change_governance_artifact_links_exact;

DROP INDEX IF EXISTS idx_change_governance_projection_snapshots_package_kind;
DROP INDEX IF EXISTS uq_change_governance_projection_snapshots_current;

DROP INDEX IF EXISTS idx_change_governance_feedback_records_package_opened;
DROP INDEX IF EXISTS idx_change_governance_feedback_records_gap_queue;

DROP INDEX IF EXISTS idx_change_governance_decision_records_package_recorded;

DROP INDEX IF EXISTS idx_change_governance_evidence_blocks_package_observed;
DROP INDEX IF EXISTS uq_change_governance_evidence_blocks_scope_kind;

DROP INDEX IF EXISTS idx_change_governance_waves_package_order;

DROP INDEX IF EXISTS idx_change_governance_internal_drafts_package_occurred;
DROP INDEX IF EXISTS uq_change_governance_internal_drafts_latest;

DROP INDEX IF EXISTS idx_change_governance_packages_high_critical_blockers;
DROP INDEX IF EXISTS idx_change_governance_packages_ready_queue;
DROP INDEX IF EXISTS idx_change_governance_packages_queue;
DROP INDEX IF EXISTS idx_change_governance_packages_latest_correlation;
DROP INDEX IF EXISTS idx_change_governance_packages_pr_number;
DROP INDEX IF EXISTS idx_change_governance_packages_repository_issue;

DROP TABLE IF EXISTS change_governance_artifact_links;
DROP TABLE IF EXISTS change_governance_projection_snapshots;
DROP TABLE IF EXISTS change_governance_feedback_records;
DROP TABLE IF EXISTS change_governance_decision_records;
DROP TABLE IF EXISTS change_governance_evidence_blocks;
DROP TABLE IF EXISTS change_governance_waves;
DROP TABLE IF EXISTS change_governance_internal_drafts;
DROP TABLE IF EXISTS change_governance_packages;
