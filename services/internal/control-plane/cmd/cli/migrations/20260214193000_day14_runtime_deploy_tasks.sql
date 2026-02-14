-- +goose Up
CREATE TABLE IF NOT EXISTS runtime_deploy_tasks (
    run_id UUID PRIMARY KEY REFERENCES agent_runs(id) ON DELETE CASCADE,
    runtime_mode TEXT NOT NULL DEFAULT '',
    namespace TEXT NOT NULL DEFAULT '',
    target_env TEXT NOT NULL DEFAULT '',
    slot_no INTEGER NOT NULL DEFAULT 0,
    repository_full_name TEXT NOT NULL DEFAULT '',
    services_yaml_path TEXT NOT NULL DEFAULT '',
    build_ref TEXT NOT NULL DEFAULT '',
    deploy_only BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL DEFAULT 'pending',
    lease_owner TEXT NULL,
    lease_until TIMESTAMPTZ NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NULL,
    result_namespace TEXT NULL,
    result_target_env TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_runtime_deploy_tasks_status CHECK (status IN ('pending', 'running', 'succeeded', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_runtime_deploy_tasks_status_lease_until
    ON runtime_deploy_tasks (status, lease_until, updated_at);

-- +goose Down
DROP INDEX IF EXISTS idx_runtime_deploy_tasks_status_lease_until;
DROP TABLE IF EXISTS runtime_deploy_tasks;
