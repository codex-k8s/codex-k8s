---
doc_id: API-CK8S-0001
type: api-contract
title: "codex-k8s — API Contract Overview"
status: draft
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# API Contract Overview: codex-k8s

## TL;DR
- Тип API: REST (public webhook + staff/private), internal gRPC (опционально).
- Аутентификация: GitHub OAuth login + short-lived JWT в API gateway + project RBAC.
- Версионирование: `/api/v1/...`.
- Основные операции: webhook ingest (public), staff/private operations для project/repository/agents/runs/slots/docs/audit и learning mode.

## Спецификации (source of truth)
- OpenAPI: `services/external/api-gateway/api/server/api.yaml`
- gRPC proto: `proto/codexk8s/v1/control_plane.proto` (будет создан)
- AsyncAPI (если есть): `services/external/api-gateway/api/server/asyncapi.yaml` (webhook/event payloads)

## Endpoints / Methods (кратко)
| Operation | Method/Topic | Path/Name | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Ingest GitHub webhook | POST | `/api/v1/webhooks/github` | signature | by delivery id | enqueue/dispatch |
| Get current user | GET | `/api/v1/me` | jwt | n/a | staff/private |
| List projects | GET | `/api/v1/projects` | jwt | n/a | RBAC-filtered, staff/private |
| Upsert project | POST | `/api/v1/projects` | jwt+admin | by project key | staff/private |
| Add project member | POST | `/api/v1/projects/{id}/members` | jwt+admin | by (project,user) | staff/private |
| Add repository | POST | `/api/v1/projects/{id}/repositories` | jwt+rw | by provider/repo | token encrypted, staff/private |
| List agents | GET | `/api/v1/agents` | jwt | n/a | fixed roster, staff/private |
| Start agent run | POST | `/api/v1/agent-runs` | jwt+rw | by correlation_id | manual trigger/override, staff/private |
| List runs | GET | `/api/v1/agent-runs` | jwt | n/a | filters/status, staff/private |
| Set learning mode | PUT | `/api/v1/projects/{id}/members/{user_id}/learning-mode` | jwt+admin | by member | toggle per user/project, staff/private |
| List learning feedback | GET | `/api/v1/agent-runs/{id}/learning-feedback` | jwt | n/a | inline + post-PR notes, staff/private |
| Update doc template | PUT | `/api/v1/docs/{doc_id}` | jwt+rw | by doc_id/version | markdown body, staff/private |
| Search docs | POST | `/api/v1/docs/search` | jwt | request hash | pgvector search, staff/private |

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

## Backward compatibility
- Что гарантируем: стабильность `/api/v1` и мягкие additive changes.
- Как деплоим изменения: staging deploy -> ручные тесты -> production gate.

## Наблюдаемость
- Логи: structured + correlation_id.
- Метрики: webhook throughput, run latency, slot usage, error rates.
- Трейсы: ingress -> domain -> db/provider/k8s.

## Решения Owner
- Для staff UI/API используется short-lived JWT через API gateway.
- Минимум public API в первой поставке: только webhook ingress.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: API границы и auth-модель MVP утверждены.
