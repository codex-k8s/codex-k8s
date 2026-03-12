---
doc_id: DM-S10-CK8S-0001
type: data-model
title: "Sprint S10 Day 5 — Data model for built-in MCP user interactions (Issue #387)"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [360, 378, 383, 385, 387, 389]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-387-data-model"
---

# Data Model: Sprint S10 built-in MCP user interactions

## TL;DR
- Schema owner остаётся `services/internal/control-plane`.
- Новые persisted сущности: `interaction_requests`, `interaction_delivery_attempts`, `interaction_callback_events`, `interaction_response_records`.
- `agent_runs` получает typed wait linkage; `agent_sessions` переиспользуется как resume snapshot storage без отдельной interaction-session таблицы.
- Главный миграционный риск: корректный backfill wait taxonomy и защита от duplicate logical completion на callback replay.

## Сущности
### Entity: `interaction_requests`
- Назначение: canonical interaction aggregate для `user.notify` и `user.decision.request`.
- Важные инварианты:
  - aggregate живёт вне approval tables;
  - decision interaction имеет не более одного effective response;
  - у одного run не более одной open interaction с `wait_reason=interaction_response`.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | interaction id |
| project_id | uuid | no |  | fk -> projects | |
| run_id | uuid | no |  | fk -> agent_runs | owner run |
| interaction_kind | text | no |  | check(notify/decision_request) | |
| state | text | no | `pending_dispatch` | check(pending_dispatch/open/resolved/expired/delivery_exhausted/cancelled) | coarse aggregate state |
| resolution_kind | text | no | `none` | check(none/delivery_only/option_selected/free_text_submitted) | final semantic outcome |
| recipient_provider | text | no |  |  | adapter routing key |
| recipient_ref | text | no |  |  | opaque destination ref |
| request_payload_json | jsonb | no | `'{}'::jsonb` |  | typed snapshot of tool input |
| context_links_json | jsonb | no | `'{}'::jsonb` |  | issue/pr/run deep-links |
| response_deadline_at | timestamptz | yes |  |  | required for decision interactions |
| effective_response_id | bigint | yes |  | fk -> interaction_response_records | logical winner |
| last_delivery_attempt_no | int | no | 0 |  | monotonic attempt counter |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: `interaction_delivery_attempts`
- Назначение: attempt-level ledger для outbound dispatch и retry behavior.
- Важные инварианты:
  - `(interaction_id, attempt_no)` уникален;
  - одновременно активен максимум один pending retry per interaction.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| interaction_id | uuid | no |  | fk -> interaction_requests | |
| attempt_no | int | no |  | unique(interaction_id, attempt_no) | |
| delivery_id | uuid | no | gen_random_uuid() | unique | dispatch idempotency key |
| adapter_kind | text | no |  |  | provider/channel family |
| status | text | no | `pending` | check(pending/accepted/delivered/failed/exhausted) | |
| request_envelope_json | jsonb | no | `'{}'::jsonb` |  | outbound snapshot |
| ack_payload_json | jsonb | yes |  |  | adapter immediate ack |
| adapter_delivery_id | text | yes |  |  | provider message id |
| retryable | bool | no | false |  | transport decision only |
| next_retry_at | timestamptz | yes |  |  | |
| last_error_code | text | yes |  |  | typed transport failure |
| started_at | timestamptz | no | now() |  | |
| finished_at | timestamptz | yes |  |  | |

### Entity: `interaction_callback_events`
- Назначение: append-only evidence для каждого callback/receipt от adapter.
- Важные инварианты:
  - `(interaction_id, adapter_event_id)` уникален;
  - raw payload сохраняется один раз и не переписывается.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| interaction_id | uuid | no |  | fk -> interaction_requests | |
| delivery_id | uuid | yes |  |  | link to attempt |
| adapter_event_id | text | no |  | unique(interaction_id, adapter_event_id) | callback dedupe key |
| callback_kind | text | no |  | check(delivery_receipt/decision_response) | |
| classification | text | no | `applied` | check(applied/duplicate/stale/expired/invalid) | domain outcome |
| normalized_payload_json | jsonb | no | `'{}'::jsonb` |  | typed DTO snapshot |
| raw_payload_json | jsonb | no | `'{}'::jsonb` |  | original adapter payload |
| received_at | timestamptz | no | now() |  | gateway receive time |
| processed_at | timestamptz | yes |  |  | |

### Entity: `interaction_response_records`
- Назначение: typed decision responses, отделённые от delivery receipts.
- Важные инварианты:
  - effective response только одна;
  - `response_kind=option` требует `selected_option_id`;
  - `response_kind=free_text` требует `free_text`.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| interaction_id | uuid | no |  | fk -> interaction_requests | |
| callback_event_id | bigint | no |  | fk -> interaction_callback_events | evidence source |
| response_kind | text | no |  | check(option/free_text) | |
| selected_option_id | text | yes |  |  | |
| free_text | text | yes |  |  | |
| responder_ref | text | yes |  |  | opaque adapter actor ref |
| classification | text | no | `applied` | check(applied/duplicate/stale/expired/invalid) | |
| is_effective | bool | no | false |  | partial unique index where true |
| responded_at | timestamptz | no | now() |  | |

### Entity: `agent_runs` (extension for interaction wait linkage)
- Назначение: coarse runtime wait-state + typed link к конкретному interaction aggregate.
- Важные инварианты:
  - `status=waiting_mcp` для interaction flow всегда сопровождается `wait_reason=interaction_response`;
  - generic `mcp` больше не используется после cutover.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| wait_reason | text | yes |  | check(owner_review/approval_pending/interaction_response) | `mcp` deprecated |
| wait_target_kind | text | yes |  | check(approval_request/interaction_request) | typed wait owner |
| wait_target_ref | text | yes |  |  | approval request id or interaction id |
| wait_deadline_at | timestamptz | yes |  |  | decision deadline for interaction waits |

### Entity: `agent_sessions` (reuse, no mandatory DDL)
- Назначение: хранить `codex-cli` session snapshot и checksum/version metadata для resume.
- Day5 decision:
  - новые interaction-specific колонки не обязательны;
  - resume payload строится из interaction aggregate и `agent_runs.wait_target_ref`, а не хранится в отдельной session table.
- Причина:
  - сохраняется один source of truth для session snapshot;
  - interaction-domain не дублирует agent session state.

### Entity: `flow_events` (payload schema hardening)
- Назначение: audit-first traceability по interaction lifecycle.
- Новые event payload keys:
  - `interaction.request.created` (`interaction_id`, `tool_name`, `run_id`)
  - `interaction.dispatch.attempted` (`interaction_id`, `attempt_no`, `delivery_id`, `adapter_kind`)
  - `interaction.callback.received` (`interaction_id`, `adapter_event_id`, `callback_kind`)
  - `interaction.response.accepted` (`interaction_id`, `response_record_id`, `response_kind`)
  - `interaction.response.rejected` (`interaction_id`, `classification`)
  - `interaction.wait.entered` (`run_id`, `interaction_id`, `wait_deadline_at`)
  - `interaction.wait.resumed` (`run_id`, `interaction_id`, `request_status`)
- Ограничение:
  - free-text response не копируется verbatim в `flow_events.payload`; туда попадают только ids/classification.

## Связи
- `agent_runs 1:N interaction_requests`
- `interaction_requests 1:N interaction_delivery_attempts`
- `interaction_requests 1:N interaction_callback_events`
- `interaction_requests 1:N interaction_response_records`
- `interaction_requests 1:0..1 effective interaction_response_records`
- `agent_runs 1:1 agent_sessions` (существующая модель, не меняется)

## Индексы и запросы (критичные)
- `interaction_requests(run_id, state, interaction_kind)`
- partial index `interaction_requests(state, response_deadline_at)` where `interaction_kind='decision_request' and state='open'`
- `interaction_delivery_attempts(interaction_id, status, next_retry_at)`
- unique index `interaction_callback_events(interaction_id, adapter_event_id)`
- partial unique index `interaction_response_records(interaction_id)` where `is_effective=true`
- `agent_runs(status, wait_reason, wait_deadline_at)` for wait queue / expiry scans

## Политика хранения данных
- `interaction_response_records.free_text` хранится только в response table и не дублируется в `flow_events`.
- `interaction_callback_events.raw_payload_json` подпадает под тот же retention policy, что и audit evidence; destructive cleanup на Day5 не вводится.
- `recipient_ref` и `responder_ref` рассматриваются как opaque operational identifiers и не показываются модели.
- Secrets не сохраняются в interaction tables; callback bearer token хранится отдельно от model-visible payload.

## Миграции (ссылка)
- См. `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`.
- Изменения `run:dev`:
  - создание interaction tables;
  - расширение `agent_runs` wait taxonomy;
  - payload hardening для `flow_events`;
  - без обязательного DDL в `agent_sessions`.

## Context7 dependency baseline
- Попытка использовать Context7 для `goose` и `kin-openapi` завершилась сообщением `Monthly quota exceeded`.
- Новых зависимостей для data-model части не требуется.

## Апрув
- request_id: owner-2026-03-12-issue-387-data-model
- Решение: pending
- Комментарий: Ожидается review interaction persistence и wait linkage.
