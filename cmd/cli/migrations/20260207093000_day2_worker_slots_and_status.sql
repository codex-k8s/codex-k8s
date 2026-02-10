-- +goose Up
ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_status;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_status
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'canceled'));

CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    slot_no INT NOT NULL,
    state TEXT NOT NULL DEFAULT 'free',
    lease_owner TEXT NULL,
    lease_until TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_slots_state CHECK (state IN ('free', 'leased', 'releasing')),
    CONSTRAINT uq_slots_project_slot UNIQUE (project_id, slot_no)
);

CREATE INDEX IF NOT EXISTS idx_slots_project_state
    ON slots (project_id, state);

CREATE INDEX IF NOT EXISTS idx_slots_state_lease_until
    ON slots (state, lease_until);

-- +goose Down
DROP INDEX IF EXISTS idx_slots_state_lease_until;
DROP INDEX IF EXISTS idx_slots_project_state;
DROP TABLE IF EXISTS slots;

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_status;
