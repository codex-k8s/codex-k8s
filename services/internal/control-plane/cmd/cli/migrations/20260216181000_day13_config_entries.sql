-- +goose Up

CREATE TABLE IF NOT EXISTS config_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope TEXT NOT NULL,
    kind TEXT NOT NULL,
    project_id UUID NULL REFERENCES projects(id) ON DELETE CASCADE,
    repository_id UUID NULL REFERENCES repositories(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value_plain TEXT NOT NULL DEFAULT '',
    value_encrypted BYTEA NULL,
    sync_targets TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    mutability TEXT NOT NULL DEFAULT 'startup_required',
    is_dangerous BOOLEAN NOT NULL DEFAULT FALSE,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    updated_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_config_entries_scope CHECK (scope IN ('platform', 'project', 'repository')),
    CONSTRAINT chk_config_entries_kind CHECK (kind IN ('secret', 'variable')),
    CONSTRAINT chk_config_entries_mutability CHECK (mutability IN ('startup_required', 'runtime_mutable')),
    CONSTRAINT chk_config_entries_scope_refs CHECK (
        (scope = 'platform' AND project_id IS NULL AND repository_id IS NULL)
        OR (scope = 'project' AND project_id IS NOT NULL AND repository_id IS NULL)
        OR (scope = 'repository' AND repository_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_config_entries_scope_key
    ON config_entries (scope, project_id, repository_id, key);

CREATE INDEX IF NOT EXISTS idx_config_entries_project_id
    ON config_entries (project_id);

CREATE INDEX IF NOT EXISTS idx_config_entries_repository_id
    ON config_entries (repository_id);

-- +goose Down
DROP INDEX IF EXISTS idx_config_entries_repository_id;
DROP INDEX IF EXISTS idx_config_entries_project_id;
DROP INDEX IF EXISTS uq_config_entries_scope_key;
DROP TABLE IF EXISTS config_entries;

