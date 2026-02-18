-- +goose Up
CREATE TABLE IF NOT EXISTS runtime_errors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    details_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    stack_trace TEXT NULL,
    correlation_id TEXT NULL,
    run_id UUID NULL REFERENCES agent_runs(id) ON DELETE SET NULL,
    project_id UUID NULL REFERENCES projects(id) ON DELETE SET NULL,
    namespace TEXT NULL,
    job_name TEXT NULL,
    viewed_at TIMESTAMPTZ NULL,
    viewed_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_runtime_errors_level CHECK (level IN ('error', 'warning', 'critical'))
);

CREATE INDEX IF NOT EXISTS idx_runtime_errors_created_at
    ON runtime_errors (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_runtime_errors_active_created_at
    ON runtime_errors (created_at DESC)
    WHERE viewed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_runtime_errors_project_created_at
    ON runtime_errors (project_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_runtime_errors_run_created_at
    ON runtime_errors (run_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_runtime_errors_run_created_at;
DROP INDEX IF EXISTS idx_runtime_errors_project_created_at;
DROP INDEX IF EXISTS idx_runtime_errors_active_created_at;
DROP INDEX IF EXISTS idx_runtime_errors_created_at;
DROP TABLE IF EXISTS runtime_errors;
