-- +goose Up

CREATE TABLE IF NOT EXISTS mission_control_entities (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    entity_kind TEXT NOT NULL,
    entity_external_key TEXT NOT NULL,
    provider_kind TEXT NOT NULL DEFAULT 'github',
    provider_url TEXT NULL,
    title TEXT NOT NULL,
    active_state TEXT NOT NULL,
    sync_status TEXT NOT NULL DEFAULT 'synced',
    projection_version BIGINT NOT NULL DEFAULT 1,
    card_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    detail_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_timeline_at TIMESTAMPTZ NULL,
    provider_updated_at TIMESTAMPTZ NULL,
    projected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    stale_after TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_mission_control_entities_kind
        CHECK (entity_kind IN ('work_item', 'discussion', 'pull_request', 'agent')),
    CONSTRAINT chk_mission_control_entities_provider_kind
        CHECK (provider_kind IN ('github', 'platform')),
    CONSTRAINT chk_mission_control_entities_active_state
        CHECK (active_state IN ('working', 'waiting', 'blocked', 'review', 'recent_critical_updates', 'archived')),
    CONSTRAINT chk_mission_control_entities_sync_status
        CHECK (sync_status IN ('synced', 'pending_sync', 'failed', 'degraded')),
    CONSTRAINT chk_mission_control_entities_projection_version
        CHECK (projection_version >= 1),
    CONSTRAINT uq_mission_control_entities_identity
        UNIQUE (project_id, entity_kind, entity_external_key),
    CONSTRAINT uq_mission_control_entities_project_row
        UNIQUE (project_id, id)
);

CREATE TABLE IF NOT EXISTS mission_control_relations (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_entity_id BIGINT NOT NULL,
    relation_kind TEXT NOT NULL,
    target_entity_id BIGINT NOT NULL,
    source_kind TEXT NOT NULL DEFAULT 'platform',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_mission_control_relations_kind
        CHECK (relation_kind IN ('linked_to', 'blocks', 'blocked_by', 'formalized_from', 'owned_by', 'assigned_to', 'tracked_by_command')),
    CONSTRAINT chk_mission_control_relations_source_kind
        CHECK (source_kind IN ('platform', 'provider', 'command', 'voice_candidate')),
    CONSTRAINT uq_mission_control_relations_edge
        UNIQUE (source_entity_id, relation_kind, target_entity_id),
    CONSTRAINT fk_mission_control_relations_source_entity
        FOREIGN KEY (project_id, source_entity_id)
        REFERENCES mission_control_entities(project_id, id)
        ON DELETE CASCADE,
    CONSTRAINT fk_mission_control_relations_target_entity
        FOREIGN KEY (project_id, target_entity_id)
        REFERENCES mission_control_entities(project_id, id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS mission_control_timeline_entries (
    id BIGSERIAL PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    entity_id BIGINT NOT NULL,
    source_kind TEXT NOT NULL,
    entry_external_key TEXT NOT NULL,
    command_id UUID NULL,
    summary TEXT NOT NULL,
    body_markdown TEXT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL,
    provider_url TEXT NULL,
    is_read_only BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_mission_control_timeline_entries_source_kind
        CHECK (source_kind IN ('provider', 'platform', 'command', 'voice_candidate')),
    CONSTRAINT uq_mission_control_timeline_entries_external_key
        UNIQUE (project_id, source_kind, entry_external_key),
    CONSTRAINT fk_mission_control_timeline_entries_entity
        FOREIGN KEY (project_id, entity_id)
        REFERENCES mission_control_entities(project_id, id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS mission_control_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    command_kind TEXT NOT NULL,
    target_entity_id BIGINT NULL,
    actor_id TEXT NOT NULL,
    business_intent_key TEXT NOT NULL,
    correlation_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'accepted',
    failure_reason TEXT NULL,
    approval_request_id UUID NULL,
    approval_state TEXT NOT NULL DEFAULT 'not_required',
    approval_requested_at TIMESTAMPTZ NULL,
    approval_decided_at TIMESTAMPTZ NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    result_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    provider_delivery_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reconciled_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_mission_control_commands_kind
        CHECK (command_kind IN ('discussion.create', 'work_item.create', 'discussion.formalize', 'stage.next_step.execute', 'command.retry_sync')),
    CONSTRAINT chk_mission_control_commands_status
        CHECK (status IN ('accepted', 'pending_approval', 'queued', 'pending_sync', 'reconciled', 'failed', 'blocked', 'cancelled')),
    CONSTRAINT chk_mission_control_commands_failure_reason
        CHECK (
            failure_reason IS NULL
            OR failure_reason IN ('provider_error', 'policy_denied', 'projection_stale', 'duplicate_intent', 'timeout', 'approval_denied', 'approval_expired', 'unknown')
        ),
    CONSTRAINT chk_mission_control_commands_approval_state
        CHECK (approval_state IN ('not_required', 'pending', 'approved', 'denied', 'expired')),
    CONSTRAINT uq_mission_control_commands_business_intent
        UNIQUE (project_id, business_intent_key),
    CONSTRAINT uq_mission_control_commands_correlation_id
        UNIQUE (correlation_id),
    CONSTRAINT uq_mission_control_commands_project_row
        UNIQUE (project_id, id),
    CONSTRAINT fk_mission_control_commands_target_entity
        FOREIGN KEY (project_id, target_entity_id)
        REFERENCES mission_control_entities(project_id, id)
        ON DELETE SET NULL (target_entity_id)
);

ALTER TABLE mission_control_timeline_entries
    DROP CONSTRAINT IF EXISTS fk_mission_control_timeline_entries_command;

ALTER TABLE mission_control_timeline_entries
    ADD CONSTRAINT fk_mission_control_timeline_entries_command
        FOREIGN KEY (project_id, command_id)
        REFERENCES mission_control_commands(project_id, id)
        ON DELETE SET NULL (command_id);

CREATE INDEX IF NOT EXISTS idx_mission_control_entities_project_active_updated
    ON mission_control_entities (project_id, active_state, projected_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_entities_project_sync_updated
    ON mission_control_entities (project_id, sync_status, projected_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_entities_project_timeline
    ON mission_control_entities (project_id, last_timeline_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_entities_card_payload
    ON mission_control_entities USING GIN (card_payload);

CREATE INDEX IF NOT EXISTS idx_mission_control_relations_source_kind
    ON mission_control_relations (source_entity_id, relation_kind, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_relations_target_kind
    ON mission_control_relations (target_entity_id, relation_kind, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_timeline_entries_entity_timeline
    ON mission_control_timeline_entries (entity_id, occurred_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_commands_project_status_updated
    ON mission_control_commands (project_id, status, updated_at DESC, requested_at DESC);

CREATE INDEX IF NOT EXISTS idx_mission_control_commands_project_approval_updated
    ON mission_control_commands (project_id, approval_state, updated_at DESC, requested_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_mission_control_commands_project_approval_updated;
DROP INDEX IF EXISTS idx_mission_control_commands_project_status_updated;
DROP INDEX IF EXISTS idx_mission_control_timeline_entries_entity_timeline;
DROP INDEX IF EXISTS idx_mission_control_relations_target_kind;
DROP INDEX IF EXISTS idx_mission_control_relations_source_kind;
DROP INDEX IF EXISTS idx_mission_control_entities_card_payload;
DROP INDEX IF EXISTS idx_mission_control_entities_project_timeline;
DROP INDEX IF EXISTS idx_mission_control_entities_project_sync_updated;
DROP INDEX IF EXISTS idx_mission_control_entities_project_active_updated;

ALTER TABLE mission_control_timeline_entries
    DROP CONSTRAINT IF EXISTS fk_mission_control_timeline_entries_command;

DROP TABLE IF EXISTS mission_control_commands;
DROP TABLE IF EXISTS mission_control_timeline_entries;
DROP TABLE IF EXISTS mission_control_relations;
DROP TABLE IF EXISTS mission_control_entities;
