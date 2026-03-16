-- +goose Up

CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value_kind TEXT NOT NULL,
    value_json JSONB NOT NULL,
    source TEXT NOT NULL,
    version BIGINT NOT NULL,
    updated_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    updated_by_email TEXT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_system_settings_value_kind CHECK (value_kind IN ('boolean')),
    CONSTRAINT chk_system_settings_source CHECK (source IN ('default', 'staff')),
    CONSTRAINT chk_system_settings_version CHECK (version > 0)
);

CREATE TABLE IF NOT EXISTS system_setting_changes (
    id BIGSERIAL PRIMARY KEY,
    setting_key TEXT NOT NULL REFERENCES system_settings(key) ON DELETE CASCADE,
    value_kind TEXT NOT NULL,
    value_json JSONB NOT NULL,
    previous_value_json JSONB NULL,
    source TEXT NOT NULL,
    version BIGINT NOT NULL,
    change_kind TEXT NOT NULL,
    actor_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    actor_email TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_system_setting_changes_value_kind CHECK (value_kind IN ('boolean')),
    CONSTRAINT chk_system_setting_changes_source CHECK (source IN ('default', 'staff')),
    CONSTRAINT chk_system_setting_changes_kind CHECK (change_kind IN ('seeded', 'updated', 'reset')),
    CONSTRAINT chk_system_setting_changes_version CHECK (version > 0),
    CONSTRAINT uq_system_setting_changes_key_version UNIQUE (setting_key, version)
);

CREATE INDEX IF NOT EXISTS idx_system_setting_changes_key_created_at
    ON system_setting_changes (setting_key, created_at DESC);

INSERT INTO system_settings (
    key,
    value_kind,
    value_json,
    source,
    version,
    updated_by_user_id,
    updated_by_email,
    updated_at
)
VALUES (
    'github_rate_limit_wait_enabled',
    'boolean',
    'false'::jsonb,
    'default',
    1,
    NULL,
    NULL,
    NOW()
)
ON CONFLICT (key) DO NOTHING;

INSERT INTO system_setting_changes (
    setting_key,
    value_kind,
    value_json,
    previous_value_json,
    source,
    version,
    change_kind,
    actor_user_id,
    actor_email,
    created_at
)
VALUES (
    'github_rate_limit_wait_enabled',
    'boolean',
    'false'::jsonb,
    NULL,
    'default',
    1,
    'seeded',
    NULL,
    NULL,
    NOW()
)
ON CONFLICT (setting_key, version) DO NOTHING;

-- +goose Down

DROP INDEX IF EXISTS idx_system_setting_changes_key_created_at;
DROP TABLE IF EXISTS system_setting_changes;
DROP TABLE IF EXISTS system_settings;
