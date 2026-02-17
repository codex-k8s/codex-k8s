-- +goose Up

ALTER TABLE repositories
    ADD COLUMN IF NOT EXISTS bot_token_encrypted BYTEA NULL,
    ADD COLUMN IF NOT EXISTS bot_username TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS bot_email TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS preflight_report_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS preflight_updated_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE repositories
    DROP COLUMN IF EXISTS preflight_updated_at,
    DROP COLUMN IF EXISTS preflight_report_json,
    DROP COLUMN IF EXISTS bot_email,
    DROP COLUMN IF EXISTS bot_username,
    DROP COLUMN IF EXISTS bot_token_encrypted;

