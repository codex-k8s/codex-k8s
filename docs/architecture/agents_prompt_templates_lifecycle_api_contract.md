---
doc_id: API-APT-CK8S-0001
type: api-contract
title: "codex-k8s — API contract: agents settings and prompt templates lifecycle"
status: in-review
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189, 195, 197]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-195-api-contract"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# API Contract: agents settings and prompt templates lifecycle

## TL;DR
- Тип API: staff/private REST (`/api/v1/staff/...`) + internal gRPC (`api-gateway -> control-plane`).
- Аутентификация: staff JWT + project RBAC.
- Версионирование: `/api/v1` для HTTP и `v1` package в proto.
- Основные операции: list/details/update для agents, lifecycle operations для prompt templates, audit history.

## Спецификации (source of truth)
- OpenAPI (to be updated in `run:dev`): `services/external/api-gateway/api/server/api.yaml`.
- gRPC proto (to be updated in `run:dev`): `proto/codexk8s/controlplane/v1/controlplane.proto`.
- Transport mapping requirements:
  - HTTP DTO: `services/external/api-gateway/internal/transport/http/models`.
  - HTTP casters: `services/external/api-gateway/internal/transport/http/casters`.
  - gRPC DTO/casters: `services/internal/control-plane/internal/transport/grpc/{models,casters}`.

## Staff HTTP endpoints (design baseline)
| Operation | Method | Path | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| List agents | GET | `/api/v1/staff/agents` | staff JWT | n/a | RBAC filtered by project |
| Get agent details | GET | `/api/v1/staff/agents/{agent_id}` | staff JWT | n/a | Includes effective policy refs |
| Update agent settings | PATCH | `/api/v1/staff/agents/{agent_id}/settings` | staff JWT + admin | `Idempotency-Key` | Partial update, optimistic validation |
| List template keys | GET | `/api/v1/staff/prompt-templates` | staff JWT | n/a | Filters: scope, role, kind, locale |
| List template versions | GET | `/api/v1/staff/prompt-templates/{template_key}/versions` | staff JWT | n/a | Sorted desc by version |
| Create template version | POST | `/api/v1/staff/prompt-templates/{template_key}/versions` | staff JWT + admin | `Idempotency-Key` | Requires `expected_version` |
| Activate template version | POST | `/api/v1/staff/prompt-templates/{template_key}/versions/{version}/activate` | staff JWT + admin | `Idempotency-Key` | Switches active version |
| Effective preview | POST | `/api/v1/staff/prompt-templates/{template_key}/preview` | staff JWT | n/a | Returns resolved content + source |
| Compare versions | GET | `/api/v1/staff/prompt-templates/{template_key}/diff` | staff JWT | n/a | Query: `from_version`, `to_version` |
| List template audit | GET | `/api/v1/staff/audit/prompt-templates` | staff JWT | n/a | Filters by project/template/actor/date |

## Internal gRPC methods (design baseline)
| RPC | Request (typed) | Response (typed) | Error mapping |
|---|---|---|---|
| `ListAgents` | `ListAgentsRequest` | `ListAgentsResponse` | `forbidden`, `internal` |
| `GetAgent` | `GetAgentRequest` | `GetAgentResponse` | `not_found`, `forbidden` |
| `UpdateAgentSettings` | `UpdateAgentSettingsRequest` | `UpdateAgentSettingsResponse` | `invalid_argument`, `conflict`, `forbidden` |
| `ListPromptTemplateVersions` | `ListPromptTemplateVersionsRequest` | `ListPromptTemplateVersionsResponse` | `not_found`, `forbidden` |
| `CreatePromptTemplateVersion` | `CreatePromptTemplateVersionRequest` | `CreatePromptTemplateVersionResponse` | `invalid_argument`, `conflict` |
| `ActivatePromptTemplateVersion` | `ActivatePromptTemplateVersionRequest` | `ActivatePromptTemplateVersionResponse` | `conflict`, `failed_precondition` |
| `PreviewPromptTemplate` | `PreviewPromptTemplateRequest` | `PreviewPromptTemplateResponse` | `failed_precondition`, `forbidden` |
| `DiffPromptTemplateVersions` | `DiffPromptTemplateVersionsRequest` | `DiffPromptTemplateVersionsResponse` | `invalid_argument`, `not_found` |
| `ListPromptTemplateAuditEvents` | `ListPromptTemplateAuditEventsRequest` | `ListPromptTemplateAuditEventsResponse` | `forbidden`, `internal` |

## DTO contract (key fields)
### PromptTemplateVersion DTO
- `template_key` (string, deterministic: `scope/role/kind/locale`)
- `version` (int32)
- `status` (`draft|active|archived`)
- `checksum` (string, sha256)
- `change_reason` (string)
- `created_by` (string)
- `created_at` (RFC3339)

### Concurrency envelope
- Request fields:
  - `expected_version` (int32, required for write).
  - `idempotency_key` (string, required for write HTTP operations).
- Conflict payload:
  - `actual_version` (int32)
  - `latest_checksum` (string)
  - `conflict_reason` (enum: `version_mismatch|active_version_changed`)

## Валидация и guardrails
- `role_key`: must be one of known system roles or registered custom role.
- `template_kind`: `work|revise` only.
- `locale`: BCP-47 formatted; minimum supported set `ru`, `en`.
- `body_markdown` size: max 128 KiB (design cap for stable diff/preview latency).
- `change_reason`: required for activate/archive operations.
- Preview request must include explicit `project_id` for project scope.

## Модель ошибок
- Canonical domain codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- Mapping rule: errors are translated only at transport boundary (HTTP error handler / gRPC interceptor).

## Retries / Rate limits
- Safe retries:
  - write operations with `Idempotency-Key`.
  - read operations (list/get/preview/diff).
- Rate limits baseline:
  - preview/diff endpoints per user/project to protect latency SLO.

## Backward compatibility
- Инициатива S6 допускает coordinated breaking changes внутри staff/private API (проект pre-production).
- Для `run:dev` требуется атомарный rollout `migrations -> internal -> edge -> frontend`, чтобы не допустить transport drift.

## Наблюдаемость
- Логи: `operation`, `project_id`, `template_key`, `version`, `status`, `correlation_id`, `duration_ms`.
- Метрики:
  - `staff_api_requests_total{endpoint,code}`
  - `prompt_template_conflict_total`
  - `prompt_template_validation_failed_total`
- Трейсы: `staff-http` span with child `control-plane-grpc` and `postgres` spans.

## Context7 dependency check
- `kin-openapi` (`/getkin/kin-openapi`): подтверждены runtime request/response validation path и поддержка OpenAPI 3.x.
- `monaco-editor` (`/microsoft/monaco-editor`): подтверждён стандартный `createDiffEditor + setModel` паттерн, достаточно для UI diff use-case.
- Вывод: для `run:dev` новые внешние библиотеки не требуются.

## Апрув
- request_id: owner-2026-02-25-issue-195-api-contract
- Решение: approved
- Комментарий: Контрактные границы зафиксированы для реализации.
