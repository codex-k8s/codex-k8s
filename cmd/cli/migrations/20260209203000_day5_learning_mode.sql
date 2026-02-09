-- +goose Up
-- Day5: learning mode settings and feedback storage.

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE project_members
    ADD COLUMN IF NOT EXISTS learning_mode_override BOOLEAN NULL;

CREATE TABLE IF NOT EXISTS learning_feedback (
    id BIGSERIAL PRIMARY KEY,
    run_id UUID NOT NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
    repository_id UUID NULL REFERENCES repositories(id) ON DELETE SET NULL,
    pr_number INT NULL,
    file_path TEXT NULL,
    line INT NULL,
    kind TEXT NOT NULL,
    explanation TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_learning_feedback_kind CHECK (kind IN ('inline', 'post_pr'))
);

CREATE INDEX IF NOT EXISTS idx_learning_feedback_run_created_at
    ON learning_feedback (run_id, created_at);

CREATE INDEX IF NOT EXISTS idx_agent_runs_project_learning_started_at
    ON agent_runs (project_id, learning_mode, started_at);

-- +goose Down
DROP INDEX IF EXISTS idx_agent_runs_project_learning_started_at;
DROP INDEX IF EXISTS idx_learning_feedback_run_created_at;
DROP TABLE IF EXISTS learning_feedback;

ALTER TABLE project_members
    DROP COLUMN IF EXISTS learning_mode_override;

ALTER TABLE projects
    DROP COLUMN IF EXISTS settings;

