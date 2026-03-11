---
doc_id: DM-S7-CK8S-0001
type: data-model
title: "Sprint S7 Day 5 — Data model for MVP readiness gap closure (Issue #238)"
status: in-review
owner_role: SA
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [212, 218, 220, 222, 238, 241]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-238-data-model"
---

# Data Model: Sprint S7 MVP readiness gap closure

## TL;DR
- Основной schema owner остаётся `services/internal/control-plane`.
- Ключевой persisted impact затрагивает `runtime_deploy_tasks`, `agent_runs`, `agent_sessions`, `flow_events`.
- Для `S7-E06`/`S7-E07` принимается policy-first подход: runtime ограничения вводятся без обязательного destructive DDL на Day5.

## Scope по потокам
| Stream | Data impact class | Решение |
|---|---|---|
| `S7-E06` | Low | физическая схема `agents/agent_policies` не меняется; deprecated поля скрываются на transport/domain write-path |
| `S7-E07` | Medium | source normalization в `prompt_templates` через доменные инварианты + optional constraint migration |
| `S7-E09` | None | UI cleanup без persisted изменений |
| `S7-E10` | High | расширение `runtime_deploy_tasks` для cancel/stop lifecycle |
| `S7-E13` | Low | stage transition evidence в `flow_events.payload` |
| `S7-E16` | Medium | терминализация `agent_runs` с явной metadata для race-safe finalization |
| `S7-E17` | High | versioned snapshot consistency в `agent_sessions` |

## Сущности
### Entity: `runtime_deploy_tasks` (extension for `S7-E10`)
- Назначение: idempotent state-machine для deploy actions `cancel/stop`.
- Важные инварианты:
  - terminal state необратим;
  - повторный `cancel/stop` после terminal state не меняет запись.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| cancel_requested_at | timestamptz | yes |  |  | additive |
| cancel_requested_by | text | yes |  |  | actor id |
| cancel_reason | text | yes |  |  | optional |
| stop_requested_at | timestamptz | yes |  |  | additive |
| stop_requested_by | text | yes |  |  | actor id |
| stop_reason | text | yes |  |  | optional |
| terminal_status_source | text | yes |  | check enum | `worker|operator|system` |
| terminal_event_seq | bigint | no | 0 |  | monotonic finalization sequence |

### Entity: `agent_runs` (extension for `S7-E16`)
- Назначение: deterministic final status при duplicate callbacks.
- Важные инварианты:
  - финализация идёт только по monotonic `terminal_event_seq`;
  - `succeeded` не понижается в `failed` при позднем duplicate callback.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| terminal_status_source | text | yes |  | check enum | `runner|worker|system` |
| terminal_event_seq | bigint | no | 0 |  | compare-and-set для finalization |
| status_reason_code | text | yes |  |  | typed reason (`duplicate_callback`, `timeout`, etc.) |

### Entity: `agent_sessions` (extension for `S7-E17`)
- Назначение: versioned snapshot read/write для self-improve reliability.
- Важные инварианты:
  - `snapshot_version` увеличивается монотонно;
  - `snapshot_checksum` соответствует текущему snapshot payload.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| snapshot_version | bigint | no | 1 |  | CAS-like upsert guard |
| snapshot_checksum | text | yes |  |  | sha256(snapshot json) |
| snapshot_updated_at | timestamptz | no | now() |  | explicit update clock |

### Entity: `flow_events` (payload schema hardening for `S7-E10/S7-E13/S7-E16/S7-E17`)
- Назначение: audit-first traceability для runtime/policy transitions.
- Новые типизированные event payload keys:
  - `runtime_deploy.cancel_requested` (`run_id`, `actor`, `reason`)
  - `runtime_deploy.stop_requested` (`run_id`, `actor`, `reason`)
  - `stage_transition.revise_selected` (`stage`, `resolver_source`, `issue_number`, `pr_number`)
  - `run.finalization_normalized` (`from_status`, `to_status`, `terminal_event_seq`)
  - `agent_session.snapshot_upserted` (`run_id`, `snapshot_version`, `snapshot_checksum`)

### Entity: `prompt_templates` (policy normalization for `S7-E07`)
- Day5 decision: без mandatory destructive DDL.
- Доменные инварианты:
  - в MVP write-path не принимает selector `repo|db`;
  - effective source фиксируется policy-driven как `repo`.
- Optional migration (run:dev): check-constraint/normalization для legacy source values.

## Связи
- `agent_runs` 1:1 `runtime_deploy_tasks` по `run_id`.
- `agent_runs` 1:1 `agent_sessions` (текущий baseline, не меняется на Day5).
- `flow_events` связывает runtime actions и stage resolver outcomes через `correlation_id`.

## Индексы и критичные запросы
- `runtime_deploy_tasks(run_id, status, updated_at)` + partial index по `status in ('pending','running')`.
- `agent_runs(correlation_id, terminal_event_seq)` для race-safe finalization.
- `agent_sessions(repository_full_name, branch_name, agent_key, snapshot_version desc)` для latest snapshot lookup.
- `flow_events(event_type, created_at desc)` + GIN по `payload` (фильтры по `run_id`/`issue_number`).

## Политика хранения данных
- `agent_sessions.codex_cli_session_json` и snapshot metadata подпадают под текущий retention policy.
- `flow_events` остаётся append-only; destructive cleanup не вводится.
- PII/secret policy без изменений: snapshot payload не должен логировать секреты.

## Domain invariants
- Terminal event sequence всегда monotonic per `run_id`.
- Repeated cancel/stop actions не создают новый terminal transition.
- Snapshot checksum должен пересчитываться на каждом успешном upsert.
- Resolver `run:<stage>:revise` для late-stage coverage публикует ровно один decision event на transition попытку.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): отсутствует.
- Migration impact (`run:dev`): additive columns + indexes для `runtime_deploy_tasks`, `agent_runs`, `agent_sessions`; payload-schema hardening для `flow_events`.

## Context7 dependency baseline
- Контрактный и UI baseline подтверждён через Context7:
  - `/getkin/kin-openapi`;
  - `/microsoft/monaco-editor`.
- Новых библиотек для data-model части не требуется.

## Апрув
- request_id: owner-2026-03-02-issue-238-data-model
- Решение: pending
- Комментарий: Ожидается review data impact и migration boundaries.
