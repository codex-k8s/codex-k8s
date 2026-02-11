---
doc_id: API-CK8S-0001
type: api-contract
title: "codex-k8s — API Contract Overview"
status: draft
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# API Contract Overview: codex-k8s

## TL;DR
- Тип API: REST (public webhook + staff/private), internal gRPC между edge и control-plane.
- Аутентификация: GitHub OAuth login + short-lived JWT в API gateway + project RBAC.
- Версионирование: `/api/v1/...`.
- Основные операции текущего среза: webhook ingest (public) + staff/private operations для auth, project/repository/user/run/learning-mode.
- Для external/staff транспорта в S2 Day1 внедрён contract-first OpenAPI (validation + backend/frontend codegen).

## Спецификации (source of truth)
- OpenAPI (api-gateway): `services/external/api-gateway/api/server/api.yaml`
- gRPC proto: `proto/codexk8s/controlplane/v1/controlplane.proto`
- AsyncAPI (если есть): `services/external/api-gateway/api/server/asyncapi.yaml` (webhook/event payloads)

## Состояние OpenAPI после S2 Day1
- OpenAPI-спека (`services/external/api-gateway/api/server/api.yaml`) покрывает все активные external/staff endpoint'ы текущего среза.
- В `api-gateway` включена runtime валидация request/response по OpenAPI (через `kin-openapi`) для `/api/*`.
- Включён backend codegen:
  - `make gen-openapi-go`
  - output: `services/external/api-gateway/internal/transport/http/generated/openapi.gen.go`
- Включён frontend codegen:
  - `make gen-openapi-ts`
  - output: `services/staff/web-console/src/shared/api/generated/**`
- В CI добавлена проверка консистентности codegen:
  - `.github/workflows/contracts_codegen_check.yml` (`make gen-openapi` + `git diff --exit-code`).

## Endpoints / Methods (текущий срез)
| Operation | Method | Path | Auth | Notes |
|---|---|---|---|---|
| Ingest GitHub webhook | POST | `/api/v1/webhooks/github` | webhook signature | idempotency по `X-GitHub-Delivery`, response status: `accepted|duplicate|ignored` |
| Start GitHub OAuth | GET | `/api/v1/auth/github/login` | public | redirect |
| Complete GitHub OAuth callback | GET | `/api/v1/auth/github/callback` | public | set auth cookie |
| Logout | POST | `/api/v1/auth/logout` | staff JWT | clears auth cookies |
| Get current principal | GET | `/api/v1/auth/me` | staff JWT | staff/private |
| List projects | GET | `/api/v1/staff/projects` | staff JWT | RBAC filtered |
| Upsert project | POST | `/api/v1/staff/projects` | staff JWT + admin | create/update by slug |
| Get project | GET | `/api/v1/staff/projects/{project_id}` | staff JWT | details |
| Delete project | DELETE | `/api/v1/staff/projects/{project_id}` | staff JWT + admin | hard delete |
| List runs | GET | `/api/v1/staff/runs` | staff JWT | run list |
| Get run | GET | `/api/v1/staff/runs/{run_id}` | staff JWT | run details |
| List run events | GET | `/api/v1/staff/runs/{run_id}/events` | staff JWT | flow events |
| List run learning feedback | GET | `/api/v1/staff/runs/{run_id}/learning-feedback` | staff JWT | educational feedback |
| List users | GET | `/api/v1/staff/users` | staff JWT | allowed users |
| Create user | POST | `/api/v1/staff/users` | staff JWT + admin | allowlist entry |
| Delete user | DELETE | `/api/v1/staff/users/{user_id}` | staff JWT + admin | remove allowlist entry |
| List project members | GET | `/api/v1/staff/projects/{project_id}/members` | staff JWT | members and roles |
| Upsert project member | POST | `/api/v1/staff/projects/{project_id}/members` | staff JWT + admin | by `user_id` or `email` |
| Delete project member | DELETE | `/api/v1/staff/projects/{project_id}/members/{user_id}` | staff JWT + admin | remove member |
| Set member learning mode override | PUT | `/api/v1/staff/projects/{project_id}/members/{user_id}/learning-mode` | staff JWT + admin | true/false/null |
| List project repositories | GET | `/api/v1/staff/projects/{project_id}/repositories` | staff JWT | repository bindings |
| Upsert project repository | POST | `/api/v1/staff/projects/{project_id}/repositories` | staff JWT + admin | token encrypted in backend |
| Delete project repository | DELETE | `/api/v1/staff/projects/{project_id}/repositories/{repository_id}` | staff JWT + admin | unbind repository |

Примечание:
- будущие маршруты (`run:*`, stage labels, prompt locale management, docs search/edit и т.д.) вводятся отдельными эпиками S2 Day2+.

## Public API boundary (MVP)
- Публично (outside/stable): только `POST /api/v1/webhooks/github`.
- Остальные endpoint'ы — staff/private API, не объявляются как public contract на первой поставке.

## Модель ошибок
- Error codes: `invalid_argument`, `unauthorized`, `forbidden`, `not_found`, `conflict`, `failed_precondition`, `internal`.
- Retries: webhook ingestion safe retry по `delivery_id`/`correlation_id`.
- Rate limits: на external webhook ingress и user API.

## Контракты данных (DTO)
- Основные сущности: user, project, project_member, repository, agent, agent_run, slot, flow_event, document.
- Валидация: schema validation + domain validation.

## Learning mode behavior
- Если learning mode активен, для user-initiated задач в prompt/context добавляется mandatory block:
  - объяснить, почему изменение сделано именно здесь;
  - какие преимущества даёт выбранный путь;
  - какие альтернативы рассмотрены и почему хуже в данном контексте.
- После создания/обновления PR worker запускает образовательный post-processing:
  - формирует комментарии по ключевым файлам и (опционально) строкам;
  - сохраняет объяснения в `learning_feedback`;
  - публикует агрегированный PR comment и, при необходимости, line-level comments.
- При выключенном learning mode pipeline работает без образовательных вставок.

## Label and stage policy behavior
- Поддерживаются классы лейблов: `run:*`, `state:*`, `need:*`.
- На текущем этапе исполнения активны триггеры `run:dev` и `run:dev:revise`; остальные `run:*` зарезервированы под поэтапное включение.
- Trigger/deploy label, инициированный агентом, проходит owner approval до применения.
- `state:*` и `need:*` могут применяться автоматически в рамках project policy.
- Любая операция с label фиксируется в `flow_events` и связывается с `agent_sessions`/`links`.

## MCP approver/executor contract behavior
- Approver/executor интеграции подключаются по HTTP-контрактам через MCP-слой.
- Telegram (`github.com/codex-k8s/telegram-approver`, `github.com/codex-k8s/telegram-executor`) рассматривается как первый адаптер контракта, но не как единственный канал.
- Контракт должен поддерживать async callbacks и единый `correlation_id` для аудита.

## Session resume and timeout behavior
- run/session поддерживает paused states `waiting_owner_review` и `waiting_mcp`.
- При `waiting_mcp` timeout-kill не применяется; таймер возобновляется после ответа MCP.
- Для resume используется сохранённый `codex-cli` session snapshot из `agent_sessions`.

## Prompt locale behavior
- Prompt templates выбираются по цепочке locale:
  - `project locale`;
  - `system default locale`;
  - fallback `en`.
- Для системных агентов baseline включает как минимум `ru` и `en` версии шаблонов.

## Backward compatibility
- Что гарантируем: стабильность `/api/v1` и мягкие additive changes.
- Как деплоим изменения: staging deploy -> ручные тесты -> production gate.

## Наблюдаемость
- Логи: structured + correlation_id.
- Метрики: webhook throughput, run latency, slot usage, label approval latency, error rates.
- Трейсы: ingress -> domain -> db/provider/k8s.

## Решения Owner
- Для staff UI/API используется short-lived JWT через API gateway.
- Минимум public API в первой поставке: только webhook ingress.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: API границы и auth-модель MVP утверждены.
