-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS agent_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id TEXT NOT NULL UNIQUE,
    project_id UUID NULL,
    agent_id UUID NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    run_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    learning_mode BOOLEAN NOT NULL DEFAULT false,
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS flow_events (
    id BIGSERIAL PRIMARY KEY,
    correlation_id TEXT NOT NULL,
    actor_type TEXT NOT NULL,
    actor_id TEXT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_runs_status_started_at
    ON agent_runs (status, started_at);

CREATE INDEX IF NOT EXISTS idx_flow_events_correlation_created_at
    ON flow_events (correlation_id, created_at);

CREATE INDEX IF NOT EXISTS idx_flow_events_event_type_created_at
    ON flow_events (event_type, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_flow_events_event_type_created_at;
DROP INDEX IF EXISTS idx_flow_events_correlation_created_at;
DROP INDEX IF EXISTS idx_agent_runs_status_started_at;
DROP TABLE IF EXISTS flow_events;
DROP TABLE IF EXISTS agent_runs;
