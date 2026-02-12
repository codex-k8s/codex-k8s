-- +goose Up

ALTER TABLE agent_sessions
    ADD COLUMN IF NOT EXISTS agent_key TEXT NOT NULL DEFAULT 'legacy-agent';

UPDATE agent_sessions
SET agent_key = 'legacy-agent'
WHERE btrim(agent_key) = '';

ALTER TABLE agent_sessions
    ALTER COLUMN agent_key DROP DEFAULT;

DROP INDEX IF EXISTS idx_agent_sessions_repo_branch_created_at;

CREATE INDEX IF NOT EXISTS idx_agent_sessions_repo_branch_agent_created_at
    ON agent_sessions (repository_full_name, branch_name, agent_key, created_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_agent_sessions_repo_branch_agent_created_at;

CREATE INDEX IF NOT EXISTS idx_agent_sessions_repo_branch_created_at
    ON agent_sessions (repository_full_name, branch_name, created_at DESC);

ALTER TABLE agent_sessions
    DROP COLUMN IF EXISTS agent_key;
