-- +goose Up

CREATE TABLE IF NOT EXISTS agent_sessions (
    id BIGSERIAL PRIMARY KEY,
    run_id UUID NOT NULL UNIQUE REFERENCES agent_runs(id) ON DELETE CASCADE,
    correlation_id TEXT NOT NULL,
    project_id UUID NULL REFERENCES projects(id) ON DELETE SET NULL,
    repository_full_name TEXT NOT NULL,
    issue_number INT NULL,
    branch_name TEXT NULL,
    pr_number INT NULL,
    pr_url TEXT NULL,
    trigger_kind TEXT NULL,
    template_kind TEXT NULL,
    template_source TEXT NULL,
    template_locale TEXT NULL,
    model TEXT NULL,
    reasoning_effort TEXT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    session_id TEXT NULL,
    session_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    codex_cli_session_path TEXT NULL,
    codex_cli_session_json JSONB NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_agent_sessions_status
        CHECK (status IN ('running', 'succeeded', 'failed', 'cancelled', 'failed_precondition'))
);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_repo_branch_created_at
    ON agent_sessions (repository_full_name, branch_name, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_issue_created_at
    ON agent_sessions (repository_full_name, issue_number, created_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_agent_sessions_issue_created_at;
DROP INDEX IF EXISTS idx_agent_sessions_repo_branch_created_at;
DROP TABLE IF EXISTS agent_sessions;
