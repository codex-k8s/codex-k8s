-- +goose Up

CREATE TABLE IF NOT EXISTS repository_preflight_locks (
    repository_id UUID PRIMARY KEY REFERENCES repositories(id) ON DELETE CASCADE,
    lock_token UUID NOT NULL,
    locked_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    locked_until TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_repository_preflight_locks_locked_until
    ON repository_preflight_locks (locked_until);

-- +goose Down
DROP INDEX IF EXISTS idx_repository_preflight_locks_locked_until;
DROP TABLE IF EXISTS repository_preflight_locks;

