---
doc_id: API-CK8S-0001
type: api-contract
title: "codex-k8s — API Contract Overview"
status: draft
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-02-13
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# API Contract Overview: codex-k8s

## TL;DR
- Тип API: REST (public webhook + staff/private), internal gRPC между edge и control-plane.
- Для agent runtime добавлен internal MCP StreamableHTTP endpoint в `control-plane` с run-bound bearer auth.
- Аутентификация: GitHub OAuth login + short-lived JWT в API gateway + project RBAC.
- Версионирование: `/api/v1/...`.
- Основные операции текущего среза: webhook ingest (public) + staff/private operations для auth, project/repository/user/run/learning-mode.
- Для external/staff транспорта в S2 Day1 внедрён contract-first OpenAPI (validation + backend/frontend codegen).
- В MVP completion (S2 Day6 + S3) добавляются API-контракты для runtime debug observability и MCP control tools orchestration.

## Спецификации (source of truth)
- OpenAPI (api-gateway): `services/external/api-gateway/api/server/api.yaml`
- gRPC proto: `proto/codexk8s/controlplane/v1/controlplane.proto`
- AsyncAPI (если есть): `services/external/api-gateway/api/server/asyncapi.yaml` (webhook/event payloads)

## Состояние MCP после S2 Day6 / S3 target
- В `control-plane` поднят MCP StreamableHTTP endpoint: `/mcp`.
- Аутентификация MCP: short-lived run-bound bearer token.
- Внутренний gRPC контракт расширен RPC `IssueRunMCPToken` для выдачи MCP токена worker-у перед запуском run pod.
- MCP-слой в текущем MVP baseline покрывает:
  - label-операции (`github_labels_*`);
  - deterministic secret sync (GitHub + Kubernetes);
  - database lifecycle (`create/delete/describe`);
  - owner feedback request (options + custom answer).
- Базовые MCP label-инструменты:
  - `github_labels_list`;
  - `github_labels_add`;
  - `github_labels_remove`;
  - `github_labels_transition` (remove+add).
- Остальные GitHub/Kubernetes runtime-операции выполняются напрямую из agent pod через `gh`/`kubectl` в рамках RBAC/policy.

## Модель доступа GitHub для агентного pod (S2 Day4)
- Агентный pod получает отдельный `CODEXK8S_GIT_BOT_TOKEN`.
- Токен используется напрямую через `gh` и `git`:
  - clone/fetch/commit/push в рабочую ветку;
  - issue/PR/review/comments операции.
- Разрешённые scopes для bot-token:
  - Read: actions, actions variables, artifact metadata, custom properties for repositories, deployments, environments, merge queues, metadata, secrets;
  - Read/Write: code, commit statuses, discussions, issues, pages, pull requests, workflows.
- Через MCP выполняются label transitions и control tools, требующие governance approvals и единый audit контур.

## Модель доступа Kubernetes для агентного pod (S2 Day4)
- Для `full-env` runner формирует `~/.kube/config` из namespaced ServiceAccount и экспортирует `KUBECONFIG`.
- Агент может выполнять через `kubectl` почти все namespaced операции runtime-диагностики и дебага.
- Исключение: прямой доступ к `secrets` (read/write) запрещён RBAC.
- Управление секретами через MCP/control-plane является частью MVP completion scope.

## Internal agent callbacks (S2 Day4)
- Для agent-runner добавлены внутренние gRPC callback RPC в `control-plane`:
  - `UpsertAgentSession` — upsert session snapshot;
  - `GetLatestAgentSession` — latest session by `(repository_full_name, branch_name, agent_key)`;
  - `InsertRunFlowEvent` — append Day4 run events.
- Авторизация callback'ов: run-bound MCP bearer token в gRPC metadata (`authorization: Bearer ...`), проверка через `VerifyRunToken`.
- Эти RPC внутренние (service-to-service), не входят в public/staff OpenAPI контракт.

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

## Endpoints / Methods (текущий и MVP target срез)
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
| List pending approvals | GET | `/api/v1/staff/approvals` | staff JWT | MCP approval queue for privileged actions |
| Resolve approval decision | POST | `/api/v1/staff/approvals/{approval_request_id}/decision` | staff JWT | approve/deny/expire/fail action request |
| List run learning feedback | GET | `/api/v1/staff/runs/{run_id}/learning-feedback` | staff JWT | educational feedback |
| List running jobs | GET | `/api/v1/staff/runs/jobs` | staff JWT | runtime jobs monitor |
| Stream run logs | GET | `/api/v1/staff/runs/{run_id}/logs/stream` | staff JWT | live tail (SSE/WebSocket) |
| List run log snapshots | GET | `/api/v1/staff/runs/{run_id}/logs` | staff JWT | historical logs |
| List wait queue | GET | `/api/v1/staff/runs/waits` | staff JWT | `waiting_mcp`/`waiting_owner_review` with reasons |
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
- маршруты staff runtime debug (`/runs/jobs`, `/runs/{run_id}/logs*`, `/runs/waits`) относятся к MVP target и вводятся в Sprint S3.
- будущие маршруты сверх MVP (`docs search/edit`, advanced policy management UI и т.д.) вводятся отдельными эпиками post-MVP.

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
- S2 baseline: активны `run:dev` и `run:dev:revise`.
- S3 target: активируется полный stage-контур `run:intake..run:ops` и `run:self-improve`.
- Trigger/deploy label, инициированный агентом, проходит owner approval до применения.
- `state:*` и `need:*` могут применяться автоматически в рамках project policy.
- Любая операция с label фиксируется в `flow_events` и связывается с `agent_sessions`/`links`.

## MCP approver/executor contract behavior
- Approver/executor интеграции подключаются по HTTP-контрактам через MCP-слой.
- Telegram (`github.com/codex-k8s/telegram-approver`, `github.com/codex-k8s/telegram-executor`) рассматривается как первый адаптер контракта, но не как единственный канал.
- Контракт должен поддерживать async callbacks и единый `correlation_id` для аудита.
- Для control tools обязателен `approval_required` режим по policy matrix.

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
