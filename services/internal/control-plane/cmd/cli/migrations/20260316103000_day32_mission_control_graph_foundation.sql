-- +goose Up

ALTER TABLE mission_control_entities
    ADD COLUMN IF NOT EXISTS continuity_status TEXT NOT NULL DEFAULT 'complete',
    ADD COLUMN IF NOT EXISTS coverage_class TEXT NOT NULL DEFAULT 'open_primary';

ALTER TABLE mission_control_entities
    DROP CONSTRAINT IF EXISTS chk_mission_control_entities_kind;

ALTER TABLE mission_control_entities
    DROP CONSTRAINT IF EXISTS chk_mission_control_entities_continuity_status;

ALTER TABLE mission_control_entities
    DROP CONSTRAINT IF EXISTS chk_mission_control_entities_coverage_class;

ALTER TABLE mission_control_entities
    ADD CONSTRAINT chk_mission_control_entities_kind
        CHECK (entity_kind IN ('work_item', 'discussion', 'pull_request', 'agent', 'run'));

ALTER TABLE mission_control_entities
    ADD CONSTRAINT chk_mission_control_entities_continuity_status
        CHECK (continuity_status IN ('complete', 'missing_run', 'missing_pull_request', 'missing_follow_up_issue', 'stale_provider', 'out_of_scope'));

ALTER TABLE mission_control_entities
    ADD CONSTRAINT chk_mission_control_entities_coverage_class
        CHECK (coverage_class IN ('open_primary', 'recent_closed_context', 'out_of_scope'));

UPDATE mission_control_entities
SET
    active_state = 'archived',
    continuity_status = 'out_of_scope',
    coverage_class = 'out_of_scope',
    updated_at = NOW()
WHERE entity_kind = 'agent'
  AND active_state <> 'archived';

ALTER TABLE mission_control_relations
    DROP CONSTRAINT IF EXISTS chk_mission_control_relations_kind;

ALTER TABLE mission_control_relations
    ADD CONSTRAINT chk_mission_control_relations_kind
        CHECK (
            relation_kind IN (
                'linked_to',
                'blocks',
                'blocked_by',
                'formalized_from',
                'owned_by',
                'assigned_to',
                'tracked_by_command',
                'spawned_run',
                'produced_pull_request',
                'continues_with',
                'related_to'
            )
        );

CREATE TABLE IF NOT EXISTS mission_control_continuity_gaps (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    subject_entity_id BIGINT NOT NULL,
    gap_kind TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'warning',
    status TEXT NOT NULL DEFAULT 'open',
    expected_entity_kind TEXT NULL,
    expected_stage_label TEXT NULL,
    resolution_entity_id BIGINT NULL,
    resolution_hint TEXT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_mission_control_continuity_gaps_kind
        CHECK (gap_kind IN ('missing_run', 'missing_pull_request', 'missing_follow_up_issue', 'provider_out_of_scope', 'provider_stale', 'orphan_node')),
    CONSTRAINT chk_mission_control_continuity_gaps_severity
        CHECK (severity IN ('blocking', 'warning', 'info')),
    CONSTRAINT chk_mission_control_continuity_gaps_status
        CHECK (status IN ('open', 'resolved', 'deferred')),
    CONSTRAINT chk_mission_control_continuity_gaps_expected_entity_kind
        CHECK (expected_entity_kind IS NULL OR expected_entity_kind IN ('discussion', 'work_item', 'run', 'pull_request')),
    CONSTRAINT fk_mission_control_continuity_gaps_subject_entity
        FOREIGN KEY (project_id, subject_entity_id)
        REFERENCES mission_control_entities(project_id, id)
        ON DELETE CASCADE,
    CONSTRAINT fk_mission_control_continuity_gaps_resolution_entity
        FOREIGN KEY (project_id, resolution_entity_id)
        REFERENCES mission_control_entities(project_id, id)
        ON DELETE SET NULL (resolution_entity_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_mission_control_continuity_gaps_open_subject_kind
    ON mission_control_continuity_gaps (project_id, subject_entity_id, gap_kind)
    WHERE status = 'open';

CREATE INDEX IF NOT EXISTS idx_mission_control_continuity_gaps_project_status_severity
    ON mission_control_continuity_gaps (project_id, status, severity, detected_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_continuity_gaps_project_subject_updated
    ON mission_control_continuity_gaps (project_id, subject_entity_id, updated_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS mission_control_workspace_watermarks (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    watermark_kind TEXT NOT NULL,
    status TEXT NOT NULL,
    summary TEXT NOT NULL,
    window_started_at TIMESTAMPTZ NULL,
    window_ended_at TIMESTAMPTZ NULL,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_mission_control_workspace_watermarks_kind
        CHECK (watermark_kind IN ('provider_freshness', 'provider_coverage', 'graph_projection', 'launch_policy')),
    CONSTRAINT chk_mission_control_workspace_watermarks_status
        CHECK (status IN ('fresh', 'stale', 'degraded', 'out_of_scope'))
);

CREATE INDEX IF NOT EXISTS idx_mission_control_workspace_watermarks_latest
    ON mission_control_workspace_watermarks (project_id, watermark_kind, observed_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_entities_project_continuity_updated
    ON mission_control_entities (project_id, active_state, continuity_status, projected_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_entities_project_coverage_updated
    ON mission_control_entities (project_id, coverage_class, projected_at DESC, id DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_mission_control_entities_project_coverage_updated;
DROP INDEX IF EXISTS idx_mission_control_entities_project_continuity_updated;
DROP INDEX IF EXISTS idx_mission_control_workspace_watermarks_latest;
DROP INDEX IF EXISTS idx_mission_control_continuity_gaps_project_subject_updated;
DROP INDEX IF EXISTS idx_mission_control_continuity_gaps_project_status_severity;
DROP INDEX IF EXISTS uq_mission_control_continuity_gaps_open_subject_kind;

DROP TABLE IF EXISTS mission_control_workspace_watermarks;
DROP TABLE IF EXISTS mission_control_continuity_gaps;

DELETE FROM mission_control_relations
WHERE relation_kind IN ('spawned_run', 'produced_pull_request', 'continues_with', 'related_to');

DELETE FROM mission_control_entities
WHERE entity_kind = 'run';

ALTER TABLE mission_control_relations
    DROP CONSTRAINT IF EXISTS chk_mission_control_relations_kind;

ALTER TABLE mission_control_relations
    ADD CONSTRAINT chk_mission_control_relations_kind
        CHECK (relation_kind IN ('linked_to', 'blocks', 'blocked_by', 'formalized_from', 'owned_by', 'assigned_to', 'tracked_by_command'));

ALTER TABLE mission_control_entities
    DROP CONSTRAINT IF EXISTS chk_mission_control_entities_coverage_class;

ALTER TABLE mission_control_entities
    DROP CONSTRAINT IF EXISTS chk_mission_control_entities_continuity_status;

ALTER TABLE mission_control_entities
    DROP CONSTRAINT IF EXISTS chk_mission_control_entities_kind;

ALTER TABLE mission_control_entities
    ADD CONSTRAINT chk_mission_control_entities_kind
        CHECK (entity_kind IN ('work_item', 'discussion', 'pull_request', 'agent'));

ALTER TABLE mission_control_entities
    DROP COLUMN IF EXISTS coverage_class,
    DROP COLUMN IF EXISTS continuity_status;
