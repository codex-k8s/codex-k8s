-- +goose Up
ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS lease_owner TEXT NULL;

ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_agent_runs_status_lease_until_started_at
    ON agent_runs (status, lease_until, started_at);

-- +goose Down
DROP INDEX IF EXISTS idx_agent_runs_status_lease_until_started_at;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS lease_until;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS lease_owner;
