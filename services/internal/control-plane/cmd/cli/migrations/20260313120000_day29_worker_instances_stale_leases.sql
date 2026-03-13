-- +goose Up
CREATE TABLE IF NOT EXISTS worker_instances (
    worker_id TEXT PRIMARY KEY,
    namespace TEXT NOT NULL DEFAULT '',
    pod_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_worker_instances_status CHECK (status IN ('active', 'stopped'))
);

CREATE INDEX IF NOT EXISTS idx_worker_instances_status_expires_at
    ON worker_instances (status, expires_at);

ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS stale_reclaim_pending BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_agent_runs_status_lease_owner
    ON agent_runs (status, lease_owner, lease_until, started_at);

-- +goose Down
DROP INDEX IF EXISTS idx_agent_runs_status_lease_owner;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS stale_reclaim_pending;

DROP INDEX IF EXISTS idx_worker_instances_status_expires_at;
DROP TABLE IF EXISTS worker_instances;
