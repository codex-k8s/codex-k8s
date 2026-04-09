---
doc_id: API-S7-CK8S-0001
type: api-contract
title: "Sprint S7 Day 5 — API contract deltas for MVP readiness gap closure (Issue #238)"
status: in-review
owner_role: SA
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [212, 218, 220, 222, 238, 241]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-238-api-contract"
---

# API Contract: Sprint S7 MVP readiness gap closure

## TL;DR
- Контрактный scope: staff/private REST (`/api/v1/staff/...`) + internal gRPC (`api-gateway -> control-plane`) + internal callbacks (`agent-runner -> control-plane`).
- Цель: зафиксировать typed contract deltas по потокам `S7-E06`, `S7-E07`, `S7-E09`, `S7-E10`, `S7-E13`, `S7-E16`, `S7-E17`.
- Общий принцип: edge остаётся thin adapter; domain semantics и state transitions определяются в `control-plane`.

## Спецификации (source of truth)
- OpenAPI: `services/external/api-gateway/api/server/api.yaml`.
- gRPC proto: `proto/kodex/controlplane/v1/controlplane.proto`.
- HTTP DTO/casters:
  - `services/external/api-gateway/internal/transport/http/models`
  - `services/external/api-gateway/internal/transport/http/casters`
- gRPC DTO/casters:
  - `services/internal/control-plane/internal/transport/grpc/{models,casters}`.

## Contract delta map by stream
| Stream | REST delta | gRPC/internal delta | Совместимость |
|---|---|---|---|
| `S7-E06` Agents settings de-scope | `PATCH /api/v1/staff/agents/{agent_id}/settings`: убрать `runtime_mode`, `prompt_locale` из write DTO | `AgentSettings` write-model сокращается до MVP-полей (`timeout_seconds`, `max_retry_count`, `approvals_required`) | coordinated breaking change допустим (pre-production) |
| `S7-E07` Prompt source repo-only | `POST /preview`/`POST /versions` и related DTO больше не принимают selector source | Prompt resolver request/response исключают dual-source flag | coordinated breaking change |
| `S7-E09` Runs UX cleanup | Контракт `DELETE /api/v1/staff/runs/{run_id}/namespace` сохраняется typed и без schema drift | `DeleteRunNamespace` RPC без breaking изменений | backward-compatible |
| `S7-E10` Runtime deploy cancel/stop | Добавить `POST /api/v1/staff/runtime-deploy/tasks/{run_id}/cancel` и `POST /api/v1/staff/runtime-deploy/tasks/{run_id}/stop` | Добавить RPC `CancelRuntimeDeployTask` / `StopRuntimeDeployTask` | additive |
| `S7-E13` multi-stage revise policy | Public REST без новых endpoints сверх next-step matrix API | Next-step action contract расширяется route для `run:doc-audit|qa|release|postdeploy|ops|self-improve:revise` | additive |
| `S7-E16` false-failed fix | Public REST без изменений | Internal finalization payload включает typed terminal metadata (`terminal_status_source`, `terminal_event_seq`) | additive |
| `S7-E17` snapshot reliability | Public REST без изменений | `UpsertAgentSession`/`GetLatestAgentSession` расширяются полями `snapshot_version`, `snapshot_checksum` | additive |

## Staff REST endpoints (target after `run:dev`)
| Operation | Method | Path | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Update agent settings (MVP fields only) | PATCH | `/api/v1/staff/agents/{agent_id}/settings` | staff JWT + admin | `Idempotency-Key` | runtime mode/locale не принимаются |
| Preview prompt template (repo-only) | POST | `/api/v1/staff/prompt-templates/{template_key}/preview` | staff JWT | n/a | source selector отсутствует |
| Force delete run namespace | DELETE | `/api/v1/staff/runs/{run_id}/namespace` | staff JWT | safe retry | existing typed contract reused |
| Cancel runtime deploy task | POST | `/api/v1/staff/runtime-deploy/tasks/{run_id}/cancel` | staff JWT + admin | `Idempotency-Key` | terminal/no-op response typed |
| Stop runtime deploy task | POST | `/api/v1/staff/runtime-deploy/tasks/{run_id}/stop` | staff JWT + admin | `Idempotency-Key` | emergency stop guardrails |

## Internal gRPC / callback methods (target after `run:dev`)
| RPC | Request delta | Response delta | Error mapping |
|---|---|---|---|
| `UpdateAgentSettings` | MVP-only settings payload | unchanged envelope | `invalid_argument`, `conflict`, `forbidden` |
| `CancelRuntimeDeployTask` (new) | `run_id`, `actor`, `reason` | `previous_status`, `current_status`, `already_terminal` | `not_found`, `failed_precondition` |
| `StopRuntimeDeployTask` (new) | `run_id`, `actor`, `reason`, `force=true` | `previous_status`, `current_status`, `already_terminal` | `invalid_argument`, `not_found`, `failed_precondition`, `forbidden` |
| `PreviewNextStepAction` / `ExecuteNextStepAction` | добавить routes `run:doc-audit|qa|release|postdeploy|ops|self-improve:revise` в resolver path | unchanged envelope | `failed_precondition` при ambiguity |
| `UpsertAgentSession` | +`snapshot_version`, +`snapshot_checksum` | +`snapshot_version` | `conflict`, `failed_precondition` |
| `GetLatestAgentSession` | optional expected checksum/version filters | +`snapshot_version`, +`snapshot_checksum` | `not_found`, `internal` |

## DTO contract decisions
### `AgentSettingsMVP` (write)
- `timeout_seconds: int32`
- `max_retry_count: int32`
- `approvals_required: bool`

### `RuntimeDeployTaskActionResponse`
- `run_id: string`
- `action: cancel|stop`
- `previous_status: string`
- `current_status: string`
- `already_terminal: bool`
- `audit_event_id: string`

### `AgentSessionSnapshot` delta
- `snapshot_version: int64`
- `snapshot_checksum: string` (sha256)

## Error contract
- Canonical codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- Stream-specific rules:
  - `S7-E13`: ambiguity всегда `failed_precondition` + `need:input` transition.
  - `S7-E16`: duplicate terminal callback не переводит run в failed, возвращает idempotent success/no-op.
  - `S7-E17`: snapshot version mismatch -> `conflict` с typed `actual_snapshot_version`.

## Retries / rate limits
- Write operations (`settings`, `cancel/stop`) требуют `Idempotency-Key`.
- Read/preview/list операции остаются safe-retry.
- Runtime action endpoints ограничиваются rate limit per `run_id`/actor для защиты от repeated destructive clicks.

## Backward compatibility
- Инициатива допускает coordinated breaking changes в staff/private API (проект pre-production).
- Для `run:dev` обязательный rollout order: `migrations -> internal -> edge -> frontend`.
- До cutover UI должен использовать generated DTO одной версии спецификации.

## Наблюдаемость
- HTTP logs: `endpoint`, `action`, `run_id`, `status_code`, `correlation_id`.
- Metrics:
  - `staff_runtime_deploy_actions_total{action,result}`
  - `stage_transition_revise_total{stage,result}`
  - `agent_settings_deprecated_field_total`
  - `agent_session_snapshot_conflict_total`
- Trace spans:
  - `staff-http -> cp-grpc -> repository` для каждого mutating action.

## Context7 verification
- `/getkin/kin-openapi`: подтверждены request/response validation APIs (`ValidateRequest`, `ValidateResponse`) для contract-first edge.
- `/microsoft/monaco-editor`: подтверждён baseline `createDiffEditor`/`setModel` для UI diff-path.
- Новые внешние зависимости для реализации контрактов S7 Day5 не требуются.

## Апрув
- request_id: owner-2026-03-02-issue-238-api-contract
- Решение: pending
- Комментарий: Ожидается review contract deltas перед handover в `run:plan`.
