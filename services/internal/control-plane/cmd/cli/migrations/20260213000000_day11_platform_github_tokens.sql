-- +goose Up
CREATE TABLE IF NOT EXISTS platform_github_tokens (
    id SMALLINT PRIMARY KEY,
    platform_token_encrypted BYTEA NULL,
    bot_token_encrypted BYTEA NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_platform_github_tokens_singleton CHECK (id = 1)
);

INSERT INTO platform_github_tokens (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS platform_github_tokens;
