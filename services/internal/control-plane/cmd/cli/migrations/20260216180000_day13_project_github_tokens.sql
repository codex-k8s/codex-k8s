-- +goose Up

CREATE TABLE IF NOT EXISTS project_github_tokens (
    project_id UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    platform_token_encrypted BYTEA NULL,
    bot_token_encrypted BYTEA NULL,
    bot_username TEXT NOT NULL DEFAULT '',
    bot_email TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS project_github_tokens;

