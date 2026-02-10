-- +goose Up

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_platform_owner BOOLEAN NOT NULL DEFAULT false;

-- Enforce "single owner" invariant at the DB level.
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_platform_owner_true
    ON users (is_platform_owner)
    WHERE is_platform_owner = true;

-- +goose Down

DROP INDEX IF EXISTS uq_users_platform_owner_true;

ALTER TABLE users
    DROP COLUMN IF EXISTS is_platform_owner;

