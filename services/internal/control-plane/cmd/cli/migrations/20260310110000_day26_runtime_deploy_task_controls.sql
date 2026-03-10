-- +goose Up
ALTER TABLE runtime_deploy_tasks
    ADD COLUMN IF NOT EXISTS cancel_requested_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS cancel_requested_by TEXT NULL,
    ADD COLUMN IF NOT EXISTS cancel_reason TEXT NULL,
    ADD COLUMN IF NOT EXISTS stop_requested_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS stop_requested_by TEXT NULL,
    ADD COLUMN IF NOT EXISTS stop_reason TEXT NULL,
    ADD COLUMN IF NOT EXISTS terminal_status_source TEXT NULL,
    ADD COLUMN IF NOT EXISTS terminal_event_seq BIGINT NOT NULL DEFAULT 0;

ALTER TABLE runtime_deploy_tasks
    DROP CONSTRAINT IF EXISTS chk_runtime_deploy_tasks_terminal_status_source;

ALTER TABLE runtime_deploy_tasks
    ADD CONSTRAINT chk_runtime_deploy_tasks_terminal_status_source
        CHECK (terminal_status_source IS NULL OR terminal_status_source IN ('worker', 'operator', 'system'));

CREATE INDEX IF NOT EXISTS idx_runtime_deploy_tasks_status_updated_at
    ON runtime_deploy_tasks (status, updated_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_runtime_deploy_tasks_status_updated_at;

ALTER TABLE runtime_deploy_tasks
    DROP CONSTRAINT IF EXISTS chk_runtime_deploy_tasks_terminal_status_source;

ALTER TABLE runtime_deploy_tasks
    DROP COLUMN IF EXISTS cancel_requested_at,
    DROP COLUMN IF EXISTS cancel_requested_by,
    DROP COLUMN IF EXISTS cancel_reason,
    DROP COLUMN IF EXISTS stop_requested_at,
    DROP COLUMN IF EXISTS stop_requested_by,
    DROP COLUMN IF EXISTS stop_reason,
    DROP COLUMN IF EXISTS terminal_status_source,
    DROP COLUMN IF EXISTS terminal_event_seq;
