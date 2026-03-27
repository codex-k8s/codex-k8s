---
doc_id: DM-S17-CK8S-0001
type: data-model
title: "Sprint S17 Day 5 — Data model for unified owner feedback loop (Issue #568)"
status: in-review
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [541, 554, 557, 559, 568, 575]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-568-data-model"
---

# Data Model: Sprint S17 unified owner feedback loop

## TL;DR
- Schema owner остаётся `services/internal/control-plane`.
- Sprint S17 не заменяет S10/S11 interaction foundation, а добавляет owner-feedback overlay для persisted request truth, wait linkage, channel projections и response bindings.
- Ключевые новые persisted streams: `owner_feedback_wait_links`, `owner_feedback_channel_projections`, `owner_feedback_response_bindings` и additive fields on `interaction_requests`, `interaction_response_records`, `agent_runs`.
- Главный миграционный риск: не допустить split-brain между Telegram и staff-console и не потерять distinction между `continuation_live` и `recovery_resume`.

## Модель расширения относительно Sprint S10/S11
- Sprint S10 remains foundation:
  - `interaction_requests`
  - `interaction_delivery_attempts`
  - `interaction_callback_events`
  - `interaction_response_records`
  - typed wait linkage in `agent_runs`
- Sprint S11 remains Telegram channel extension:
  - `interaction_channel_bindings`
  - `interaction_callback_handles`
  - Telegram evidence/provider refs
- Sprint S17 adds owner-feedback specialization:
  - canonical owner-feedback status on `interaction_requests`
  - explicit wait linkage and recovery classification
  - dual-surface projections for Telegram/staff-console parity
  - one response binding registry across Telegram and staff-console

## Сущности
### Entity: `interaction_requests` (S17 extension)
- Назначение: сохранять canonical owner-feedback truth поверх existing interaction aggregate.
- Важные инварианты:
  - `interaction_family=owner_feedback` означает same-session-first continuation semantics;
  - canonical status остаётся единственным источником owner/operator visibility;
  - Telegram and staff-console read the same row-level status and deadlines.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `interaction_family` | text | no | `generic` | check(generic/owner_feedback) | S17 overlay flag |
| `canonical_status` | text | no | `created` | check(created/delivery_pending/delivery_accepted/waiting/response_received/continuation_live/recovery_resume/overdue/expired/manual_fallback/resolved) | product-visible status |
| `primary_surface` | text | no | `telegram` | check(telegram/staff_console) | current primary owner-facing surface |
| `staff_fallback_enabled` | bool | no | true |  | explicit staff-console availability |
| `live_continuation_required` | bool | no | false |  | true for owner feedback waits |
| `owner_wait_deadline_at` | timestamptz | yes |  |  | hard response deadline |
| `overdue_at` | timestamptz | yes |  |  | soft visibility threshold |
| `response_source_kind` | text | yes |  | check(telegram_callback/telegram_free_text/telegram_voice/staff_option/staff_free_text) | accepted source |

### Entity: `owner_feedback_wait_links`
- Назначение: typed join between request aggregate, run wait-state, session heartbeat and recovery-only resume.
- Важные инварианты:
  - one open wait link per owner-feedback interaction;
  - `run_id` is unique while request remains open;
  - `continuation_path=live_same_session` may switch to `recovery_resume` only after live-session loss evidence.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `interaction_id` | uuid | no |  | pk, fk -> interaction_requests | one wait link per request |
| `run_id` | uuid | no |  | unique, fk -> agent_runs | owner run |
| `owner_wait_window_seconds` | int | no |  | check(>0) | logical wait window |
| `live_wait_deadline_at` | timestamptz | no |  |  | hard deadline aligned with request |
| `tool_timeout_deadline_at` | timestamptz | no |  |  | effective MCP wait timeout/TTL |
| `wait_started_at` | timestamptz | no | now() |  | wait entered timestamp |
| `last_session_heartbeat_at` | timestamptz | yes |  |  | copied from `agent_sessions` evidence |
| `latest_snapshot_version` | bigint | yes |  |  | last persisted snapshot version |
| `recovery_state` | text | no | `live_expected` | check(live_expected/live_lost/recovery_pending/recovery_resumed/recovery_failed) | explicit degraded classification |
| `continuation_path` | text | no | `live_same_session` | check(live_same_session/recovery_resume/manual_fallback/expired) | typed continuation outcome |
| `created_at` | timestamptz | no | now() |  | |
| `updated_at` | timestamptz | no | now() |  | |

### Entity: `owner_feedback_channel_projections`
- Назначение: persisted read-model for Telegram and staff-console parity without giving UI/adapter ownership of semantics.
- Важные инварианты:
  - multiple projections may exist, but all reflect one `interaction_requests.canonical_status`;
  - staff-console projection is not a delivery binding and does not own retries;
  - stale projections remain visible as stale evidence, not silent overwrite.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `id` | bigserial | no |  | pk | |
| `interaction_id` | uuid | no |  | fk -> interaction_requests | |
| `projection_kind` | text | no |  | check(telegram_owner_inbox/staff_owner/staff_operator) | surface family |
| `projection_state` | text | no | `pending_publish` | check(pending_publish/visible/overdue/expired/manual_fallback/stale/closed) | read-model state |
| `source_delivery_attempt_id` | bigint | yes |  | fk -> interaction_delivery_attempts | Telegram evidence when present |
| `source_channel_binding_id` | bigint | yes |  | fk -> interaction_channel_bindings | Telegram binding link |
| `presentation_ref_json` | jsonb | no | `'{}'::jsonb` |  | UI/provider evidence only |
| `last_rendered_status` | text | no | `created` |  | copy of canonical status at render time |
| `visible_since` | timestamptz | yes |  |  | |
| `stale_at` | timestamptz | yes |  |  | |
| `closed_at` | timestamptz | yes |  |  | |
| `created_at` | timestamptz | no | now() |  | |
| `updated_at` | timestamptz | no | now() |  | |

### Entity: `owner_feedback_response_bindings`
- Назначение: one logical binding registry for Telegram option handle, Telegram free-text session and staff-console bindings.
- Важные инварианты:
  - one binding resolves to exactly one request;
  - only one effective response may consume a binding;
  - Telegram external correlators are stored only as hashes.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `id` | uuid | no | gen_random_uuid() | pk | canonical binding id |
| `interaction_id` | uuid | no |  | fk -> interaction_requests | |
| `projection_id` | bigint | yes |  | fk -> owner_feedback_channel_projections | surface that exposed binding |
| `binding_kind` | text | no |  | check(telegram_callback_option/telegram_free_text_session/staff_option/staff_free_text) | closed variant |
| `option_id` | text | yes |  |  | required for option bindings |
| `external_ref_hash` | bytea | yes |  | unique | Telegram handle hash or signed UI token hash |
| `state` | text | no | `open` | check(open/applied/duplicate/stale/expired/revoked) | binding lifecycle |
| `expires_at` | timestamptz | no |  |  | business deadline |
| `grace_expires_at` | timestamptz | no |  |  | post-deadline classification window |
| `accepted_response_id` | bigint | yes |  | fk -> interaction_response_records | effective response link |
| `created_at` | timestamptz | no | now() |  | |
| `updated_at` | timestamptz | no | now() |  | |

### Entity: `interaction_response_records` (S17 extension)
- Назначение: хранить accepted and rejected owner responses with channel-neutral classification.
- Важные инварианты:
  - one effective response per request remains enforced by partial unique index;
  - voice is normalized to text but keeps separate `response_source_kind`.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `response_source_kind` | text | no |  | check(telegram_callback/telegram_free_text/telegram_voice/staff_option/staff_free_text) | final source |
| `response_binding_id` | uuid | yes |  | fk -> owner_feedback_response_bindings | |
| `normalized_text` | text | yes |  |  | free-text or voice transcript |

### Entity: `agent_runs` (S17 extension)
- Назначение: сохранять coarse runtime wait-state with owner-feedback continuation hints.
- Важные инварианты:
  - `status=waiting_mcp` and `wait_reason=interaction_response` remain unchanged for owner feedback;
  - `continuation_path` is additive runtime hint, not replacement for canonical request status.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `continuation_path` | text | yes |  | check(live_same_session/recovery_resume/manual_fallback/expired) | runtime-visible path |
| `wait_started_at` | timestamptz | yes |  |  | coarse wait queue field |

### Entity: `agent_sessions` (reuse, no mandatory S17 DDL)
- Назначение: хранить `codex-cli` session snapshot, heartbeat and timeout-guard evidence.
- S17 decision:
  - no mandatory new columns;
  - linkage goes through `owner_feedback_wait_links.run_id -> agent_runs.id` and existing `agent_sessions.run_id`;
  - `snapshot_version`, `snapshot_checksum`, `last_heartbeat_at`, `wait_state`, `timeout_guard_disabled` remain sufficient.

### Entity: `flow_events` (payload extension)
- Назначение: audit-first traceability for owner-feedback lifecycle.
- Новые event payload keys:
  - `owner_feedback.request.created`
  - `owner_feedback.delivery.accepted`
  - `owner_feedback.response.classified`
  - `owner_feedback.visibility.transitioned`
  - `owner_feedback.continuation.live`
  - `owner_feedback.continuation.recovery`
  - `owner_feedback.wait.overdue`
  - `owner_feedback.wait.expired`
  - `owner_feedback.manual_fallback.raised`
- Ограничение:
  - free-text and voice transcript do not get copied verbatim into `flow_events.payload`.

## Связи
- `interaction_requests 1:1 owner_feedback_wait_links` for `interaction_family=owner_feedback`
- `interaction_requests 1:N owner_feedback_channel_projections`
- `interaction_requests 1:N owner_feedback_response_bindings`
- `owner_feedback_channel_projections 1:N owner_feedback_response_bindings`
- `owner_feedback_response_bindings 1:0..1 interaction_response_records`
- `agent_runs 1:1 owner_feedback_wait_links` while request is open
- `agent_runs 1:1 agent_sessions` (existing baseline, reused)

## Индексы и запросы (критичные)
- Query: open waits by deadline / overdue state
  - index on `owner_feedback_wait_links(live_wait_deadline_at, recovery_state)`
- Query: staff console pending list
  - index on `owner_feedback_channel_projections(projection_kind, projection_state, updated_at desc)`
- Query: response binding lookup by opaque handle
  - unique index on `owner_feedback_response_bindings(external_ref_hash)` where `external_ref_hash is not null`
- Query: accepted response by request
  - partial unique index on `interaction_response_records(interaction_id)` where `is_effective=true`
- Query: runtime wait queue
  - index on `agent_runs(status, wait_reason, continuation_path, wait_started_at)`

## Политика хранения данных
- Free text and voice transcripts remain only in `interaction_response_records.normalized_text`.
- `presentation_ref_json` and Telegram provider refs are operational evidence and must not be surfaced to model-visible output.
- Staff-console bindings do not store raw UI tokens; only hashed external refs are persisted.
- Historical S10/S11 generic interactions are not backfilled into owner-feedback overlays.

## Доменные инварианты
- `interaction_family=owner_feedback` implies `live_continuation_required=true`.
- `manual_fallback` cannot be cleared by UI-only write; it changes only through domain decision.
- `recovery_resume` can occur only after `recovery_state=live_lost`.
- `expired` requests may still record callback evidence inside grace window, but cannot produce new effective responses.
- Telegram and staff-console always resolve to the same `interaction_requests.id`.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): absent, docs only.
- Migration impact (`run:dev`):
  - additive overlay tables and columns;
  - no schema ownership change;
  - no destructive rewrite of S10/S11 evidence tables.

## Миграции (ссылка)
- See `docs/architecture/initiatives/s17_unified_owner_feedback_loop/migrations_policy.md`.
- Sprint S10/S11 interaction foundation remains a hard prerequisite.

## Context7 / dependencies
- Новые внешние зависимости на этапе `run:design` не выбирались и не добавлялись.
- Context7 не требовался, потому что пакет фиксирует contracts и persisted model поверх уже утверждённой Sprint S10/S11 foundation.

## Апрув
- request_id: owner-2026-03-27-issue-568-data-model
- Решение: pending
- Комментарий: Ожидается review owner-feedback overlays, wait linkage и projection model.
