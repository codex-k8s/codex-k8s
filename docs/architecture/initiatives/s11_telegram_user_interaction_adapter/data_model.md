---
doc_id: DM-S11-CK8S-0001
type: data-model
title: "Sprint S11 Day 5 — Data model for Telegram user interaction adapter (Issue #454)"
status: in-review
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-454-data-model"
---

# Data Model: Sprint S11 Telegram user interaction adapter

## TL;DR
- Schema owner остаётся `services/internal/control-plane`.
- Sprint S11 не заменяет S10 interaction schema, а расширяет её Telegram-specific persisted state.
- Ключевые новые сущности и расширения: `interaction_channel_bindings`, `interaction_callback_handles`, extensions for `interaction_requests`, `interaction_delivery_attempts`, `interaction_callback_events`.
- Главный миграционный риск: корректное разделение business state, provider refs и operator visibility без хранения raw callback handles/tokens в plaintext.

## Модель расширения относительно Sprint S10
- Sprint S10 остаётся foundation:
  - `interaction_requests`
  - `interaction_delivery_attempts`
  - `interaction_callback_events`
  - `interaction_response_records`
  - `agent_runs` typed wait linkage
- Sprint S11 adds only Telegram channel extension data:
  - channel binding and provider message refs
  - hashed callback handles
  - operator visibility fields for Telegram continuation/fallback
  - delivery-role split for `dispatch|edit|follow_up`

## Сущности
### Entity: `interaction_channel_bindings`
- Назначение: persisted binding между interaction aggregate и channel-specific provider message refs/continuation state.
- Важные инварианты:
  - одна active binding запись на `(interaction_id, adapter_kind)`;
  - raw callback bearer token не хранится в таблице;
  - provider refs считаются opaque operational identifiers и не попадают в agent-facing payload.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| interaction_id | uuid | no |  | fk -> interaction_requests | |
| adapter_kind | text | no |  | check(telegram) | first channel-specific binding |
| recipient_ref | text | no |  |  | opaque platform recipient routing ref |
| provider_chat_ref | text | yes |  |  | opaque Telegram chat ref |
| provider_message_ref_json | jsonb | no | `'{}'::jsonb` |  | typed provider message snapshot (`message_id`, `inline_message_id`, `sent_at`) |
| callback_token_key_id | text | yes |  |  | key/version id, not raw token |
| callback_token_expires_at | timestamptz | yes |  |  | delivery token grace deadline |
| edit_capability | text | no | `unknown` | check(unknown/editable/keyboard_only/follow_up_only) | adapter ack result |
| continuation_state | text | no | `pending_primary_delivery` | check(pending_primary_delivery/ready_for_edit/follow_up_required/manual_fallback_required/closed) | |
| last_operator_signal_code | text | yes |  | check(delivery_retry_exhausted/invalid_callback_payload/expired_wait/edit_fallback_sent/follow_up_failed/manual_resume_required) | latest visible signal |
| last_operator_signal_at | timestamptz | yes |  |  | |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: `interaction_callback_handles`
- Назначение: secure lookup table for Telegram callback and free-text session handles.
- Важные инварианты:
  - raw handle string в БД не хранится, только `handle_hash`;
  - один handle соответствует ровно одному semantic target;
  - handle может быть использован максимум один раз как effective response.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| interaction_id | uuid | no |  | fk -> interaction_requests | |
| channel_binding_id | bigint | no |  | fk -> interaction_channel_bindings | |
| handle_hash | bytea | no |  | unique | sha256(raw_handle) |
| handle_kind | text | no |  | check(option/free_text_session) | |
| option_id | text | yes |  |  | required for `option` |
| state | text | no | `open` | check(open/used/expired/revoked) | |
| response_deadline_at | timestamptz | no |  |  | business deadline |
| grace_expires_at | timestamptz | no |  |  | deadline + 24h |
| used_callback_event_id | bigint | yes |  | fk -> interaction_callback_events | |
| used_at | timestamptz | yes |  |  | |
| created_at | timestamptz | no | now() |  | |

### Entity: `interaction_requests` (S11 extension)
- Назначение: сохранять channel family и operator visibility поверх S10 aggregate.
- Важные инварианты:
  - channel-specific extension не меняет core semantic state machine;
  - operator-visible state не заменяет audit trail, а дополняет его.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| channel_family | text | no | `platform_only` | check(platform_only/telegram) | additive S11 field |
| active_channel_binding_id | bigint | yes |  | fk -> interaction_channel_bindings | current delivery binding |
| operator_state | text | no | `nominal` | check(nominal/watch/manual_fallback_required/resolved) | platform read model |
| operator_signal_code | text | yes |  | check(delivery_retry_exhausted/invalid_callback_payload/expired_wait/edit_fallback_sent/follow_up_failed/manual_resume_required) | current visible reason |
| operator_signal_at | timestamptz | yes |  |  | |

### Entity: `interaction_delivery_attempts` (S11 extension)
- Назначение: отделить primary delivery, message edit и follow-up notify within one interaction.
- Важные инварианты:
  - `(interaction_id, attempt_no)` остаётся уникальным;
  - continuation attempts are explicit by `delivery_role`;
  - retry logic differs by role but keeps one ledger.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| channel_binding_id | bigint | yes |  | fk -> interaction_channel_bindings | |
| delivery_role | text | no | `primary_dispatch` | check(primary_dispatch/message_edit/follow_up_notify) | |
| provider_message_ref_json | jsonb | no | `'{}'::jsonb` |  | snapshot used for edit/follow-up |
| continuation_reason | text | yes |  | check(applied_response/edit_failed/expired_wait/operator_fallback) | |

### Entity: `interaction_callback_events` (S11 extension)
- Назначение: хранить Telegram-specific evidence without moving semantics into provider payload.
- Важные инварианты:
  - `(interaction_id, adapter_event_id)` remains unique;
  - handle reference is stored as hash, not raw handle;
  - raw provider identifiers are evidence-only.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| channel_binding_id | bigint | yes |  | fk -> interaction_channel_bindings | |
| callback_handle_hash | bytea | yes |  |  | sha256(raw_handle) |
| provider_message_ref_json | jsonb | no | `'{}'::jsonb` |  | |
| provider_update_id | text | yes |  |  | Telegram update id as opaque evidence |
| provider_callback_query_id | text | yes |  |  | callback query id as opaque evidence |

### Entity: `interaction_response_records` (S11 extension, optional columns)
- Назначение: связать effective response с конкретным channel binding и handle kind.
- Важные инварианты:
  - effective response remains unique per interaction;
  - free-text is stored only here, not in operator summary fields.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| channel_binding_id | bigint | yes |  | fk -> interaction_channel_bindings | source binding |
| handle_kind | text | yes |  | check(option/free_text_session) | |

### Entity: `agent_runs` (reuse, no new mandatory S11 columns)
- Назначение: keep coarse wait-state and resume linkage from Sprint S10.
- S11 decision:
  - no new Telegram-specific columns are required;
  - `wait_target_kind=interaction_request` and `wait_target_ref=<interaction_id>` remain sufficient;
  - Telegram-specific continuation state stays in interaction/channel tables, not in run aggregate.

## Связи
- `interaction_requests 1:N interaction_channel_bindings`
- `interaction_channel_bindings 1:N interaction_callback_handles`
- `interaction_channel_bindings 1:N interaction_delivery_attempts`
- `interaction_channel_bindings 1:N interaction_callback_events`
- `interaction_requests 1:N interaction_response_records`
- `interaction_requests 1:0..1 active_channel_binding_id`

## Индексы и запросы (критичные)
- Query: resolve callback handle
  - unique index on `interaction_callback_handles(handle_hash)`
- Query: active Telegram interactions requiring operator attention
  - index on `interaction_requests(channel_family, operator_state, operator_signal_at desc)`
- Query: ready continuation after applied callback
  - index on `interaction_channel_bindings(continuation_state, updated_at)`
- Query: edit/follow-up attempts per interaction
  - index on `interaction_delivery_attempts(interaction_id, delivery_role, next_retry_at)`
- Query: provider evidence lookup
  - index on `interaction_callback_events(channel_binding_id, processed_at desc)`
- Query: provider message ref uniqueness
  - partial unique index on `(adapter_kind, provider_chat_ref, (provider_message_ref_json->>'message_id'))`
    where `provider_message_ref_json ? 'message_id'`

## Политика хранения данных
- `callback_handle_hash` persists as long as interaction audit evidence persists; raw handles are never stored.
- `provider_message_ref_json` is operational evidence and should not be exposed to model-visible output.
- `free_text` remains only in `interaction_response_records` and never duplicates into `flow_events.payload`.
- `operator_state` and `operator_signal_code` are projection fields; historical detail remains in `flow_events`.
- Raw callback bearer token is not persisted.

## Доменные инварианты
- Only one effective response may move an interaction from open to terminal.
- `manual_fallback_required` cannot be cleared without either successful follow-up continuation or explicit operator action recorded in `flow_events`.
- `edit_capability=follow_up_only` forbids scheduling `delivery_role=message_edit`.
- A used or expired callback handle can still classify late callbacks, but can never produce a new effective response.
- `channel_family=telegram` requires an `interaction_channel_bindings` row before primary dispatch is acknowledged.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): absent, docs only.
- Migration impact (`run:dev`):
  - additive Telegram extension tables and columns on top of S10 interaction schema;
  - hash-based callback lookup and new indexes;
  - no schema ownership changes outside `control-plane`.

## Миграции (ссылка)
- See `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/migrations_policy.md`.
- S11 migrations must run only after S10 interaction foundation migrations are available.

## Открытые вопросы
- Нужно ли materialize operator visibility as a dedicated read-model table later, если `interaction_requests.operator_state` станет bottleneck для staff UI?
- Нужно ли разносить notify receipts and decision callbacks по разным partial indexes already in first implementation wave?

## Апрув
- request_id: owner-2026-03-14-issue-454-data-model
- Решение: pending
- Комментарий: Ожидается review Telegram extension schema, operator visibility fields и handle-hash persistence.
