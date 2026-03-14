-- +goose Up

CREATE TABLE IF NOT EXISTS github_rate_limit_waits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    run_id UUID NOT NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
    contour_kind TEXT NOT NULL,
    signal_origin TEXT NOT NULL,
    operation_class TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'open',
    limit_kind TEXT NOT NULL,
    confidence TEXT NOT NULL DEFAULT 'deterministic',
    recovery_hint_kind TEXT NOT NULL,
    dominant_for_run BOOLEAN NOT NULL DEFAULT false,
    signal_id TEXT NOT NULL,
    request_fingerprint TEXT NULL,
    correlation_id TEXT NOT NULL,
    resume_action_kind TEXT NOT NULL,
    resume_payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    manual_action_kind TEXT NULL,
    auto_resume_attempts_used INT NOT NULL DEFAULT 0,
    max_auto_resume_attempts INT NOT NULL DEFAULT 0,
    resume_not_before TIMESTAMPTZ NULL,
    last_resume_attempt_at TIMESTAMPTZ NULL,
    first_detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_signal_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_github_rate_limit_waits_contour_kind
        CHECK (contour_kind IN ('platform_pat', 'agent_bot_token')),
    CONSTRAINT chk_github_rate_limit_waits_signal_origin
        CHECK (signal_origin IN ('control_plane', 'worker', 'agent_runner')),
    CONSTRAINT chk_github_rate_limit_waits_operation_class
        CHECK (operation_class IN ('run_status_comment', 'issue_label_transition', 'repository_provider_call', 'agent_github_call')),
    CONSTRAINT chk_github_rate_limit_waits_state
        CHECK (state IN ('open', 'auto_resume_scheduled', 'auto_resume_in_progress', 'resolved', 'manual_action_required', 'cancelled')),
    CONSTRAINT chk_github_rate_limit_waits_limit_kind
        CHECK (limit_kind IN ('primary', 'secondary')),
    CONSTRAINT chk_github_rate_limit_waits_confidence
        CHECK (confidence IN ('deterministic', 'conservative', 'provider_uncertain')),
    CONSTRAINT chk_github_rate_limit_waits_recovery_hint_kind
        CHECK (recovery_hint_kind IN ('rate_limit_reset', 'retry_after', 'exponential_backoff', 'manual_only')),
    CONSTRAINT chk_github_rate_limit_waits_resume_action_kind
        CHECK (resume_action_kind IN ('run_status_comment_retry', 'platform_github_call_replay', 'agent_session_resume')),
    CONSTRAINT chk_github_rate_limit_waits_manual_action_kind
        CHECK (manual_action_kind IS NULL OR manual_action_kind IN ('requeue_platform_operation', 'resume_agent_session', 'retry_after_operator_review')),
    CONSTRAINT chk_github_rate_limit_waits_signal_id
        UNIQUE (signal_id)
);

CREATE TABLE IF NOT EXISTS github_rate_limit_wait_evidence (
    id BIGSERIAL PRIMARY KEY,
    wait_id UUID NOT NULL REFERENCES github_rate_limit_waits(id) ON DELETE CASCADE,
    event_kind TEXT NOT NULL,
    signal_id TEXT NULL,
    signal_origin TEXT NULL,
    provider_status_code INT NULL,
    retry_after_seconds INT NULL,
    rate_limit_limit INT NULL,
    rate_limit_remaining INT NULL,
    rate_limit_used INT NULL,
    rate_limit_reset_at TIMESTAMPTZ NULL,
    rate_limit_resource TEXT NULL,
    github_request_id TEXT NULL,
    documentation_url TEXT NULL,
    message_excerpt TEXT NULL,
    stderr_excerpt TEXT NULL,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_github_rate_limit_wait_evidence_event_kind
        CHECK (event_kind IN ('signal_detected', 'classified', 'resume_scheduled', 'resume_attempted', 'resume_failed', 'resolved', 'manual_action_required', 'comment_mirror_failed')),
    CONSTRAINT chk_github_rate_limit_wait_evidence_signal_origin
        CHECK (signal_origin IS NULL OR signal_origin IN ('control_plane', 'worker', 'agent_runner'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_github_rate_limit_waits_open_run_contour
    ON github_rate_limit_waits (run_id, contour_kind)
    WHERE state IN ('open', 'auto_resume_scheduled', 'auto_resume_in_progress', 'manual_action_required');

CREATE UNIQUE INDEX IF NOT EXISTS uq_github_rate_limit_waits_open_dominant_per_run
    ON github_rate_limit_waits (run_id)
    WHERE dominant_for_run = true
      AND state IN ('open', 'auto_resume_scheduled', 'auto_resume_in_progress', 'manual_action_required');

CREATE INDEX IF NOT EXISTS idx_github_rate_limit_waits_resume_queue
    ON github_rate_limit_waits (state, resume_not_before)
    WHERE state IN ('open', 'auto_resume_scheduled');

CREATE INDEX IF NOT EXISTS idx_github_rate_limit_waits_project_queue
    ON github_rate_limit_waits (project_id, state, dominant_for_run, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_github_rate_limit_waits_correlation_id
    ON github_rate_limit_waits (correlation_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_github_rate_limit_wait_evidence_wait_event_signal
    ON github_rate_limit_wait_evidence (wait_id, event_kind, signal_id)
    WHERE signal_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_github_rate_limit_wait_evidence_wait_observed_at
    ON github_rate_limit_wait_evidence (wait_id, observed_at DESC, id DESC);

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_status;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_status
        CHECK (status IN ('pending', 'running', 'waiting_backpressure', 'succeeded', 'failed', 'canceled'));

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_wait_reason;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_wait_reason
        CHECK (wait_reason IS NULL OR wait_reason IN ('owner_review', 'approval_pending', 'interaction_response', 'github_rate_limit'));

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_wait_target_kind;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_wait_target_kind
        CHECK (wait_target_kind IS NULL OR wait_target_kind IN ('approval_request', 'interaction_request', 'github_rate_limit_wait'));

ALTER TABLE agent_sessions
    DROP CONSTRAINT IF EXISTS chk_agent_sessions_wait_state;

ALTER TABLE agent_sessions
    ADD CONSTRAINT chk_agent_sessions_wait_state
        CHECK (wait_state IS NULL OR wait_state IN ('owner_review', 'mcp', 'backpressure'));

-- +goose Down

ALTER TABLE agent_sessions
    DROP CONSTRAINT IF EXISTS chk_agent_sessions_wait_state;

ALTER TABLE agent_sessions
    ADD CONSTRAINT chk_agent_sessions_wait_state
        CHECK (wait_state IS NULL OR wait_state IN ('owner_review', 'mcp'));

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_wait_target_kind;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_wait_target_kind
        CHECK (wait_target_kind IS NULL OR wait_target_kind IN ('approval_request', 'interaction_request'));

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_wait_reason;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_wait_reason
        CHECK (wait_reason IS NULL OR wait_reason IN ('owner_review', 'approval_pending', 'interaction_response'));

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS chk_agent_runs_status;

ALTER TABLE agent_runs
    ADD CONSTRAINT chk_agent_runs_status
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'canceled'));

DROP INDEX IF EXISTS idx_github_rate_limit_wait_evidence_wait_observed_at;
DROP INDEX IF EXISTS uq_github_rate_limit_wait_evidence_wait_event_signal;
DROP INDEX IF EXISTS idx_github_rate_limit_waits_correlation_id;
DROP INDEX IF EXISTS idx_github_rate_limit_waits_project_queue;
DROP INDEX IF EXISTS idx_github_rate_limit_waits_resume_queue;
DROP INDEX IF EXISTS uq_github_rate_limit_waits_open_dominant_per_run;
DROP INDEX IF EXISTS uq_github_rate_limit_waits_open_run_contour;

DROP TABLE IF EXISTS github_rate_limit_wait_evidence;
DROP TABLE IF EXISTS github_rate_limit_waits;
