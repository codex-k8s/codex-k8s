-- +goose Up
CREATE TABLE IF NOT EXISTS run_access_keys (
    run_id UUID PRIMARY KEY REFERENCES agent_runs(id) ON DELETE CASCADE,
    project_id UUID NULL,
    correlation_id TEXT NOT NULL,
    runtime_mode TEXT NOT NULL DEFAULT 'code-only',
    namespace TEXT NULL,
    target_env TEXT NULL,
    key_hash BYTEA NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    last_used_at TIMESTAMPTZ NULL,
    created_by TEXT NOT NULL DEFAULT 'system',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_run_access_keys_status CHECK (status IN ('active', 'revoked'))
);

CREATE INDEX IF NOT EXISTS idx_run_access_keys_status_expires_at
    ON run_access_keys (status, expires_at);

CREATE INDEX IF NOT EXISTS idx_run_access_keys_last_used_at
    ON run_access_keys (last_used_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_run_access_keys_last_used_at;
DROP INDEX IF EXISTS idx_run_access_keys_status_expires_at;
DROP TABLE IF EXISTS run_access_keys;
