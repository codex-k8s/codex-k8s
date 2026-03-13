-- +goose Up

CREATE TABLE IF NOT EXISTS interaction_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    run_id UUID NOT NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
    interaction_kind TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'pending_dispatch',
    resolution_kind TEXT NOT NULL DEFAULT 'none',
    recipient_provider TEXT NOT NULL,
    recipient_ref TEXT NOT NULL,
    request_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    context_links_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_deadline_at TIMESTAMPTZ NULL,
    effective_response_id BIGINT NULL,
    last_delivery_attempt_no INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_interaction_requests_kind
        CHECK (interaction_kind IN ('notify', 'decision_request')),
    CONSTRAINT chk_interaction_requests_state
        CHECK (state IN ('pending_dispatch', 'open', 'resolved', 'expired', 'delivery_exhausted', 'cancelled')),
    CONSTRAINT chk_interaction_requests_resolution_kind
        CHECK (resolution_kind IN ('none', 'delivery_only', 'option_selected', 'free_text_submitted'))
);

CREATE TABLE IF NOT EXISTS interaction_delivery_attempts (
    id BIGSERIAL PRIMARY KEY,
    interaction_id UUID NOT NULL REFERENCES interaction_requests(id) ON DELETE CASCADE,
    attempt_no INT NOT NULL,
    delivery_id UUID NOT NULL DEFAULT gen_random_uuid(),
    adapter_kind TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    request_envelope_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    ack_payload_json JSONB NULL,
    adapter_delivery_id TEXT NULL,
    retryable BOOLEAN NOT NULL DEFAULT false,
    next_retry_at TIMESTAMPTZ NULL,
    last_error_code TEXT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_interaction_delivery_attempts_status
        CHECK (status IN ('pending', 'accepted', 'delivered', 'failed', 'exhausted')),
    CONSTRAINT uq_interaction_delivery_attempts_attempt_no
        UNIQUE (interaction_id, attempt_no),
    CONSTRAINT uq_interaction_delivery_attempts_delivery_id
        UNIQUE (delivery_id)
);

CREATE TABLE IF NOT EXISTS interaction_callback_events (
    id BIGSERIAL PRIMARY KEY,
    interaction_id UUID NOT NULL REFERENCES interaction_requests(id) ON DELETE CASCADE,
    delivery_id UUID NULL,
    adapter_event_id TEXT NOT NULL,
    callback_kind TEXT NOT NULL,
    classification TEXT NOT NULL DEFAULT 'applied',
    normalized_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    raw_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_interaction_callback_events_kind
        CHECK (callback_kind IN ('delivery_receipt', 'decision_response')),
    CONSTRAINT chk_interaction_callback_events_classification
        CHECK (classification IN ('applied', 'duplicate', 'stale', 'expired', 'invalid')),
    CONSTRAINT uq_interaction_callback_events_adapter_event
        UNIQUE (interaction_id, adapter_event_id),
    CONSTRAINT fk_interaction_callback_events_delivery_id
        FOREIGN KEY (delivery_id) REFERENCES interaction_delivery_attempts(delivery_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS interaction_response_records (
    id BIGSERIAL PRIMARY KEY,
    interaction_id UUID NOT NULL REFERENCES interaction_requests(id) ON DELETE CASCADE,
    callback_event_id BIGINT NOT NULL REFERENCES interaction_callback_events(id) ON DELETE CASCADE,
    response_kind TEXT NOT NULL,
    selected_option_id TEXT NULL,
    free_text TEXT NULL,
    responder_ref TEXT NULL,
    classification TEXT NOT NULL DEFAULT 'applied',
    is_effective BOOLEAN NOT NULL DEFAULT false,
    responded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_interaction_response_records_kind
        CHECK (response_kind IN ('option', 'free_text')),
    CONSTRAINT chk_interaction_response_records_classification
        CHECK (classification IN ('applied', 'duplicate', 'stale', 'expired', 'invalid')),
    CONSTRAINT chk_interaction_response_records_payload
        CHECK (
            (response_kind = 'option' AND selected_option_id IS NOT NULL AND COALESCE(free_text, '') = '')
            OR (response_kind = 'free_text' AND COALESCE(selected_option_id, '') = '' AND free_text IS NOT NULL)
        )
);

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS fk_interaction_requests_effective_response_id;

ALTER TABLE interaction_requests
    ADD CONSTRAINT fk_interaction_requests_effective_response_id
        FOREIGN KEY (effective_response_id) REFERENCES interaction_response_records(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_interaction_requests_run_state_kind
    ON interaction_requests (run_id, state, interaction_kind);

CREATE UNIQUE INDEX IF NOT EXISTS uq_interaction_requests_open_decision_per_run
    ON interaction_requests (run_id)
    WHERE interaction_kind = 'decision_request'
      AND state IN ('pending_dispatch', 'open');

CREATE INDEX IF NOT EXISTS idx_interaction_requests_open_decision_deadline
    ON interaction_requests (state, response_deadline_at)
    WHERE interaction_kind = 'decision_request'
      AND state IN ('pending_dispatch', 'open');

CREATE INDEX IF NOT EXISTS idx_interaction_delivery_attempts_interaction_status_retry
    ON interaction_delivery_attempts (interaction_id, status, next_retry_at);

CREATE INDEX IF NOT EXISTS idx_interaction_callback_events_interaction_delivery
    ON interaction_callback_events (interaction_id, delivery_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_interaction_response_records_effective
    ON interaction_response_records (interaction_id)
    WHERE is_effective = true;

ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS wait_reason TEXT NULL;

ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS wait_target_kind TEXT NULL;

ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS wait_target_ref TEXT NULL;

ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS wait_deadline_at TIMESTAMPTZ NULL;

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_wait_reason;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_wait_reason
        CHECK (wait_reason IS NULL OR wait_reason IN ('owner_review', 'approval_pending', 'interaction_response'));

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_wait_target_kind;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_wait_target_kind
        CHECK (wait_target_kind IS NULL OR wait_target_kind IN ('approval_request', 'interaction_request'));

CREATE INDEX IF NOT EXISTS idx_agent_runs_status_wait_reason_deadline_at
    ON agent_runs (status, wait_reason, wait_deadline_at);

WITH latest_sessions AS (
    SELECT DISTINCT ON (ags.run_id)
        ags.run_id,
        COALESCE(ags.wait_state, '') AS wait_state
    FROM agent_sessions ags
    WHERE COALESCE(ags.wait_state, '') IN ('mcp', 'owner_review')
    ORDER BY ags.run_id, ags.updated_at DESC, ags.created_at DESC
),
latest_approval_requests AS (
    SELECT DISTINCT ON (mar.run_id)
        mar.run_id,
        mar.id AS approval_request_id
    FROM mcp_action_requests mar
    WHERE mar.run_id IS NOT NULL
      AND mar.approval_state IN ('requested', 'approved')
    ORDER BY mar.run_id, mar.created_at DESC
)
UPDATE agent_runs ar
SET
    wait_reason = CASE
        WHEN COALESCE(ar.wait_reason, '') <> '' THEN ar.wait_reason
        WHEN ls.wait_state = 'mcp' THEN 'approval_pending'
        WHEN ls.wait_state = 'owner_review' THEN 'owner_review'
        ELSE NULL
    END,
    wait_target_kind = CASE
        WHEN COALESCE(ar.wait_target_kind, '') <> '' THEN ar.wait_target_kind
        WHEN ls.wait_state = 'mcp' AND lar.approval_request_id IS NOT NULL THEN 'approval_request'
        ELSE NULL
    END,
    wait_target_ref = CASE
        WHEN COALESCE(ar.wait_target_ref, '') <> '' THEN ar.wait_target_ref
        WHEN ls.wait_state = 'mcp' AND lar.approval_request_id IS NOT NULL THEN lar.approval_request_id::text
        ELSE NULL
    END,
    updated_at = NOW()
FROM latest_sessions ls
LEFT JOIN latest_approval_requests lar ON lar.run_id = ls.run_id
WHERE ar.id = ls.run_id;

-- +goose Down

DROP INDEX IF EXISTS idx_agent_runs_status_wait_reason_deadline_at;

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_wait_target_kind;

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_wait_reason;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS wait_deadline_at;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS wait_target_ref;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS wait_target_kind;

ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS wait_reason;

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS fk_interaction_requests_effective_response_id;

DROP INDEX IF EXISTS uq_interaction_response_records_effective;
DROP INDEX IF EXISTS idx_interaction_callback_events_interaction_delivery;
DROP INDEX IF EXISTS idx_interaction_delivery_attempts_interaction_status_retry;
DROP INDEX IF EXISTS idx_interaction_requests_open_decision_deadline;
DROP INDEX IF EXISTS uq_interaction_requests_open_decision_per_run;
DROP INDEX IF EXISTS idx_interaction_requests_run_state_kind;

DROP TABLE IF EXISTS interaction_response_records;
DROP TABLE IF EXISTS interaction_callback_events;
DROP TABLE IF EXISTS interaction_delivery_attempts;
DROP TABLE IF EXISTS interaction_requests;
