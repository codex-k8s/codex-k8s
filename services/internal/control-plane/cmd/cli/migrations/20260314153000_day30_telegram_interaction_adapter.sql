-- +goose Up

CREATE TABLE IF NOT EXISTS interaction_channel_bindings (
    id BIGSERIAL PRIMARY KEY,
    interaction_id UUID NOT NULL REFERENCES interaction_requests(id) ON DELETE CASCADE,
    adapter_kind TEXT NOT NULL,
    recipient_ref TEXT NOT NULL,
    provider_chat_ref TEXT NULL,
    provider_message_ref_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    callback_token_key_id TEXT NULL,
    callback_token_expires_at TIMESTAMPTZ NULL,
    edit_capability TEXT NOT NULL DEFAULT 'unknown',
    continuation_state TEXT NOT NULL DEFAULT 'pending_primary_delivery',
    last_operator_signal_code TEXT NULL,
    last_operator_signal_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_interaction_channel_bindings_adapter_kind
        CHECK (adapter_kind IN ('telegram')),
    CONSTRAINT chk_interaction_channel_bindings_edit_capability
        CHECK (edit_capability IN ('unknown', 'editable', 'keyboard_only', 'follow_up_only')),
    CONSTRAINT chk_interaction_channel_bindings_continuation_state
        CHECK (continuation_state IN ('pending_primary_delivery', 'ready_for_edit', 'follow_up_required', 'manual_fallback_required', 'closed')),
    CONSTRAINT chk_interaction_channel_bindings_operator_signal
        CHECK (
            last_operator_signal_code IS NULL
            OR last_operator_signal_code IN (
                'delivery_retry_exhausted',
                'invalid_callback_payload',
                'expired_wait',
                'edit_fallback_sent',
                'follow_up_failed',
                'manual_resume_required'
            )
        ),
    CONSTRAINT uq_interaction_channel_bindings_active
        UNIQUE (interaction_id, adapter_kind)
);

CREATE TABLE IF NOT EXISTS interaction_callback_handles (
    id BIGSERIAL PRIMARY KEY,
    interaction_id UUID NOT NULL REFERENCES interaction_requests(id) ON DELETE CASCADE,
    channel_binding_id BIGINT NOT NULL REFERENCES interaction_channel_bindings(id) ON DELETE CASCADE,
    handle_hash BYTEA NOT NULL,
    handle_kind TEXT NOT NULL,
    option_id TEXT NULL,
    state TEXT NOT NULL DEFAULT 'open',
    response_deadline_at TIMESTAMPTZ NOT NULL,
    grace_expires_at TIMESTAMPTZ NOT NULL,
    used_callback_event_id BIGINT NULL REFERENCES interaction_callback_events(id) ON DELETE SET NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_interaction_callback_handles_kind
        CHECK (handle_kind IN ('option', 'free_text_session')),
    CONSTRAINT chk_interaction_callback_handles_state
        CHECK (state IN ('open', 'used', 'expired', 'revoked')),
    CONSTRAINT uq_interaction_callback_handles_hash
        UNIQUE (handle_hash)
);

ALTER TABLE interaction_requests
    ADD COLUMN IF NOT EXISTS channel_family TEXT NOT NULL DEFAULT 'platform_only';

ALTER TABLE interaction_requests
    ADD COLUMN IF NOT EXISTS active_channel_binding_id BIGINT NULL;

ALTER TABLE interaction_requests
    ADD COLUMN IF NOT EXISTS operator_state TEXT NOT NULL DEFAULT 'nominal';

ALTER TABLE interaction_requests
    ADD COLUMN IF NOT EXISTS operator_signal_code TEXT NULL;

ALTER TABLE interaction_requests
    ADD COLUMN IF NOT EXISTS operator_signal_at TIMESTAMPTZ NULL;

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS chk_interaction_requests_channel_family;

ALTER TABLE interaction_requests
    ADD CONSTRAINT chk_interaction_requests_channel_family
        CHECK (channel_family IN ('platform_only', 'telegram'));

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS chk_interaction_requests_operator_state;

ALTER TABLE interaction_requests
    ADD CONSTRAINT chk_interaction_requests_operator_state
        CHECK (operator_state IN ('nominal', 'watch', 'manual_fallback_required', 'resolved'));

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS chk_interaction_requests_operator_signal_code;

ALTER TABLE interaction_requests
    ADD CONSTRAINT chk_interaction_requests_operator_signal_code
        CHECK (
            operator_signal_code IS NULL
            OR operator_signal_code IN (
                'delivery_retry_exhausted',
                'invalid_callback_payload',
                'expired_wait',
                'edit_fallback_sent',
                'follow_up_failed',
                'manual_resume_required'
            )
        );

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS fk_interaction_requests_active_channel_binding_id;

ALTER TABLE interaction_requests
    ADD CONSTRAINT fk_interaction_requests_active_channel_binding_id
        FOREIGN KEY (active_channel_binding_id) REFERENCES interaction_channel_bindings(id) ON DELETE SET NULL;

ALTER TABLE interaction_delivery_attempts
    ADD COLUMN IF NOT EXISTS channel_binding_id BIGINT NULL REFERENCES interaction_channel_bindings(id) ON DELETE SET NULL;

ALTER TABLE interaction_delivery_attempts
    ADD COLUMN IF NOT EXISTS delivery_role TEXT NOT NULL DEFAULT 'primary_dispatch';

ALTER TABLE interaction_delivery_attempts
    ADD COLUMN IF NOT EXISTS provider_message_ref_json JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE interaction_delivery_attempts
    ADD COLUMN IF NOT EXISTS continuation_reason TEXT NULL;

ALTER TABLE interaction_delivery_attempts
    DROP CONSTRAINT IF EXISTS chk_interaction_delivery_attempts_role;

ALTER TABLE interaction_delivery_attempts
    ADD CONSTRAINT chk_interaction_delivery_attempts_role
        CHECK (delivery_role IN ('primary_dispatch', 'message_edit', 'follow_up_notify'));

ALTER TABLE interaction_delivery_attempts
    DROP CONSTRAINT IF EXISTS chk_interaction_delivery_attempts_continuation_reason;

ALTER TABLE interaction_delivery_attempts
    ADD CONSTRAINT chk_interaction_delivery_attempts_continuation_reason
        CHECK (
            continuation_reason IS NULL
            OR continuation_reason IN ('applied_response', 'edit_failed', 'expired_wait', 'operator_fallback')
        );

ALTER TABLE interaction_callback_events
    ADD COLUMN IF NOT EXISTS channel_binding_id BIGINT NULL REFERENCES interaction_channel_bindings(id) ON DELETE SET NULL;

ALTER TABLE interaction_callback_events
    ADD COLUMN IF NOT EXISTS callback_handle_hash BYTEA NULL;

ALTER TABLE interaction_callback_events
    ADD COLUMN IF NOT EXISTS provider_message_ref_json JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE interaction_callback_events
    ADD COLUMN IF NOT EXISTS provider_update_id TEXT NULL;

ALTER TABLE interaction_callback_events
    ADD COLUMN IF NOT EXISTS provider_callback_query_id TEXT NULL;

ALTER TABLE interaction_callback_events
    DROP CONSTRAINT IF EXISTS chk_interaction_callback_events_kind;

ALTER TABLE interaction_callback_events
    ADD CONSTRAINT chk_interaction_callback_events_kind
        CHECK (callback_kind IN ('delivery_receipt', 'option_selected', 'free_text_received', 'transport_failure'));

ALTER TABLE interaction_response_records
    ADD COLUMN IF NOT EXISTS channel_binding_id BIGINT NULL REFERENCES interaction_channel_bindings(id) ON DELETE SET NULL;

ALTER TABLE interaction_response_records
    ADD COLUMN IF NOT EXISTS handle_kind TEXT NULL;

ALTER TABLE interaction_response_records
    DROP CONSTRAINT IF EXISTS chk_interaction_response_records_handle_kind;

ALTER TABLE interaction_response_records
    ADD CONSTRAINT chk_interaction_response_records_handle_kind
        CHECK (handle_kind IS NULL OR handle_kind IN ('option', 'free_text_session'));

CREATE INDEX IF NOT EXISTS idx_interaction_channel_bindings_state_updated_at
    ON interaction_channel_bindings (continuation_state, updated_at);

CREATE INDEX IF NOT EXISTS idx_interaction_requests_operator_state_signal_at
    ON interaction_requests (channel_family, operator_state, operator_signal_at DESC);

CREATE INDEX IF NOT EXISTS idx_interaction_delivery_attempts_role_retry
    ON interaction_delivery_attempts (interaction_id, delivery_role, next_retry_at);

CREATE INDEX IF NOT EXISTS idx_interaction_callback_events_binding_processed_at
    ON interaction_callback_events (channel_binding_id, processed_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_interaction_channel_bindings_provider_message
    ON interaction_channel_bindings (adapter_kind, provider_chat_ref, ((provider_message_ref_json->>'message_id')))
    WHERE provider_message_ref_json ? 'message_id';

-- +goose Down

DROP INDEX IF EXISTS uq_interaction_channel_bindings_provider_message;
DROP INDEX IF EXISTS idx_interaction_callback_events_binding_processed_at;
DROP INDEX IF EXISTS idx_interaction_delivery_attempts_role_retry;
DROP INDEX IF EXISTS idx_interaction_requests_operator_state_signal_at;
DROP INDEX IF EXISTS idx_interaction_channel_bindings_state_updated_at;

ALTER TABLE interaction_response_records
    DROP CONSTRAINT IF EXISTS chk_interaction_response_records_handle_kind;

ALTER TABLE interaction_response_records
    DROP COLUMN IF EXISTS handle_kind;

ALTER TABLE interaction_response_records
    DROP COLUMN IF EXISTS channel_binding_id;

ALTER TABLE interaction_callback_events
    DROP CONSTRAINT IF EXISTS chk_interaction_callback_events_kind;

ALTER TABLE interaction_callback_events
    ADD CONSTRAINT chk_interaction_callback_events_kind
        CHECK (callback_kind IN ('delivery_receipt', 'decision_response'));

ALTER TABLE interaction_callback_events
    DROP COLUMN IF EXISTS provider_callback_query_id;

ALTER TABLE interaction_callback_events
    DROP COLUMN IF EXISTS provider_update_id;

ALTER TABLE interaction_callback_events
    DROP COLUMN IF EXISTS provider_message_ref_json;

ALTER TABLE interaction_callback_events
    DROP COLUMN IF EXISTS callback_handle_hash;

ALTER TABLE interaction_callback_events
    DROP COLUMN IF EXISTS channel_binding_id;

ALTER TABLE interaction_delivery_attempts
    DROP CONSTRAINT IF EXISTS chk_interaction_delivery_attempts_continuation_reason;

ALTER TABLE interaction_delivery_attempts
    DROP CONSTRAINT IF EXISTS chk_interaction_delivery_attempts_role;

ALTER TABLE interaction_delivery_attempts
    DROP COLUMN IF EXISTS continuation_reason;

ALTER TABLE interaction_delivery_attempts
    DROP COLUMN IF EXISTS provider_message_ref_json;

ALTER TABLE interaction_delivery_attempts
    DROP COLUMN IF EXISTS delivery_role;

ALTER TABLE interaction_delivery_attempts
    DROP COLUMN IF EXISTS channel_binding_id;

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS fk_interaction_requests_active_channel_binding_id;

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS chk_interaction_requests_operator_signal_code;

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS chk_interaction_requests_operator_state;

ALTER TABLE interaction_requests
    DROP CONSTRAINT IF EXISTS chk_interaction_requests_channel_family;

ALTER TABLE interaction_requests
    DROP COLUMN IF EXISTS operator_signal_at;

ALTER TABLE interaction_requests
    DROP COLUMN IF EXISTS operator_signal_code;

ALTER TABLE interaction_requests
    DROP COLUMN IF EXISTS operator_state;

ALTER TABLE interaction_requests
    DROP COLUMN IF EXISTS active_channel_binding_id;

ALTER TABLE interaction_requests
    DROP COLUMN IF EXISTS channel_family;

DROP TABLE IF EXISTS interaction_callback_handles;
DROP TABLE IF EXISTS interaction_channel_bindings;
