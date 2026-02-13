-- +goose Up

CREATE TABLE IF NOT EXISTS mcp_action_requests (
    id BIGSERIAL PRIMARY KEY,
    correlation_id TEXT NOT NULL,
    run_id UUID NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
    tool_name TEXT NOT NULL,
    action TEXT NOT NULL,
    target_ref JSONB NOT NULL DEFAULT '{}'::jsonb,
    approval_mode TEXT NOT NULL DEFAULT 'owner',
    approval_state TEXT NOT NULL DEFAULT 'requested',
    requested_by TEXT NOT NULL,
    applied_by TEXT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_mcp_action_requests_approval_mode
        CHECK (approval_mode IN ('none', 'owner', 'delegated')),
    CONSTRAINT chk_mcp_action_requests_approval_state
        CHECK (approval_state IN ('requested', 'approved', 'denied', 'expired', 'failed', 'applied'))
);

CREATE INDEX IF NOT EXISTS idx_mcp_action_requests_state_created_at
    ON mcp_action_requests (approval_state, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_mcp_action_requests_correlation_created_at
    ON mcp_action_requests (correlation_id, created_at DESC);

ALTER TABLE agent_sessions
    ADD COLUMN IF NOT EXISTS wait_state TEXT NULL DEFAULT NULL;

ALTER TABLE agent_sessions
    ADD COLUMN IF NOT EXISTS timeout_guard_disabled BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE agent_sessions
    ADD COLUMN IF NOT EXISTS last_heartbeat_at TIMESTAMPTZ NULL;

ALTER TABLE agent_sessions
    DROP CONSTRAINT IF EXISTS chk_agent_sessions_wait_state;

ALTER TABLE agent_sessions
    ADD CONSTRAINT chk_agent_sessions_wait_state
        CHECK (wait_state IS NULL OR wait_state IN ('owner_review', 'mcp'));

CREATE INDEX IF NOT EXISTS idx_agent_sessions_run_wait_state_heartbeat
    ON agent_sessions (run_id, wait_state, last_heartbeat_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_agent_sessions_run_wait_state_heartbeat;

ALTER TABLE agent_sessions
    DROP CONSTRAINT IF EXISTS chk_agent_sessions_wait_state;

ALTER TABLE agent_sessions
    DROP COLUMN IF EXISTS last_heartbeat_at;

ALTER TABLE agent_sessions
    DROP COLUMN IF EXISTS timeout_guard_disabled;

ALTER TABLE agent_sessions
    DROP COLUMN IF EXISTS wait_state;

DROP INDEX IF EXISTS idx_mcp_action_requests_correlation_created_at;
DROP INDEX IF EXISTS idx_mcp_action_requests_state_created_at;
DROP TABLE IF EXISTS mcp_action_requests;
