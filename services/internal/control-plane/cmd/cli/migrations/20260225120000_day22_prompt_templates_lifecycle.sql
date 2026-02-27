-- +goose Up
ALTER TABLE agents
    ADD COLUMN IF NOT EXISTS settings_version INTEGER NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS prompt_templates (
    id BIGSERIAL PRIMARY KEY,
    scope_type TEXT NOT NULL,
    scope_id UUID NULL REFERENCES projects(id) ON DELETE CASCADE,
    role_key TEXT NOT NULL,
    template_kind TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT 'en',
    body_markdown TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'global_override',
    version INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL DEFAULT 'draft',
    checksum TEXT NOT NULL,
    change_reason TEXT NULL,
    supersedes_version INTEGER NULL,
    updated_by TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMPTZ NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_prompt_templates_scope_type CHECK (scope_type IN ('global', 'project')),
    CONSTRAINT chk_prompt_templates_template_kind CHECK (template_kind IN ('work', 'revise')),
    CONSTRAINT chk_prompt_templates_source CHECK (source IN ('project_override', 'global_override', 'repo_seed')),
    CONSTRAINT chk_prompt_templates_status CHECK (status IN ('draft', 'active', 'archived')),
    CONSTRAINT chk_prompt_templates_scope_id CHECK (
        (scope_type = 'global' AND scope_id IS NULL)
        OR (scope_type = 'project' AND scope_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_prompt_templates_scope_version
    ON prompt_templates (scope_type, COALESCE(scope_id::text, ''), role_key, template_kind, locale, version);

CREATE UNIQUE INDEX IF NOT EXISTS uq_prompt_templates_active_version
    ON prompt_templates (scope_type, COALESCE(scope_id::text, ''), role_key, template_kind, locale)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_prompt_templates_scope_version_desc
    ON prompt_templates (scope_type, COALESCE(scope_id::text, ''), role_key, template_kind, locale, version DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_prompt_templates_scope_version_desc;
DROP INDEX IF EXISTS uq_prompt_templates_active_version;
DROP INDEX IF EXISTS uq_prompt_templates_scope_version;
DROP TABLE IF EXISTS prompt_templates;

ALTER TABLE agents
    DROP COLUMN IF EXISTS settings_version;

