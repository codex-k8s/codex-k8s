---
doc_id: EPC-CK8S-S1-D1
type: epic
title: "Epic Day 1: Webhook ingress and idempotency"
status: completed
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-day1"
---

# Epic Day 1: Webhook ingress and idempotency

## TL;DR
- Цель эпика: стабильно принимать GitHub webhook и создавать `agent_runs` без дублей.
- Ключевая ценность: предсказуемый входной контур webhook-driven модели.
- MVP-результат: рабочий `POST /api/v1/webhooks/github` с signature validation и dedup.

## Priority
- `P0` (входной контур платформы).

## Ожидаемые артефакты дня
- Реализация webhook ingress и signature verification в `services/external/api-gateway/**`.
- Реализация dedup/idempotency policy и run bootstrap state в доменном слое `services/external/api-gateway/internal/domain/**`.
- Миграция/DDL обновления для уникальности `correlation_id` и индексов `flow_events`.
- Smoke evidence webhook replay на staging и обновление документации контракта API.

## Контекст
- Почему эпик нужен: это единственный публичный API для MVP.
- Связь с требованиями: FR-003, FR-025, NFR-003.

## Scope
### In scope
- Endpoint webhook ingress в `api-gateway`.
- Проверка подписи вебхука.
- Идемпотентность по delivery id/correlation id.
- Запись событий в `flow_events` и стартовых записей в `agent_runs`.

### Out of scope
- Полный набор staff/private endpoints.
- Поддержка провайдеров кроме GitHub.

## Декомпозиция (Stories/Tasks)
- Story-1: HTTP handler и валидация payload.
- Story-2: signature verification и ошибки.
- Story-3: dedup policy и state transition `pending`.
- Story-4: smoke tests на staging webhook flow.

## Data model impact (по шаблону data_model.md)
- Сущности:
  - `agent_runs`: использование `correlation_id` как уникального ключа обработки.
  - `flow_events`: append-only запись webhook ingress событий.
- Связи/FK:
  - В Day 1 `project_id`/`agent_id` остаются незаполненными (будут заполнены в Day 2+ при orchestration mapping).
- Индексы и запросы:
  - Проверить наличие/создать индекс `agent_runs(correlation_id)` (unique).
  - Проверить наличие/создать индекс `flow_events(correlation_id, created_at)`.
- Миграции:
  - Добавить миграции только если индексы/ограничения отсутствуют.
- Retention/PII:
  - В payload не хранить секреты подписи, только безопасные поля контекста.

## Критерии приемки эпика
- Повторная доставка одного webhook не создаёт второй run.
- Ошибочная подпись отклоняется.
- После merge изменения задеплоены на staging и проверены вручную.

## Evidence
- `POST /api/v1/webhooks/github` реализован в `services/external/api-gateway/internal/transport/http/webhook_handler.go`.
- Валидация подписи реализована в `libs/go/crypto/githubsignature/verify.go`.
- Idempotency и запись `agent_runs`/`flow_events` реализованы в:
  - `services/external/api-gateway/internal/domain/webhook/service.go`
  - `services/external/api-gateway/internal/repository/postgres/agentrun/repository.go`
  - `services/external/api-gateway/internal/repository/postgres/flowevent/repository.go`
- DDL миграция добавлена:
  - `services/internal/control-plane/cmd/cli/migrations/20260206191000_day1_webhook_ingest.sql`
- Контракт OpenAPI/AsyncAPI добавлен:
  - `services/external/api-gateway/api/server/api.yaml`
  - `services/external/api-gateway/api/server/asyncapi.yaml`
- Unit tests:
  - `libs/go/crypto/githubsignature/verify_test.go`
  - `services/external/api-gateway/internal/domain/webhook/service_test.go`
  - `services/external/api-gateway/internal/transport/http/webhook_handler_test.go`
- Verification commands:
  - `go test ./...`
  - `bash -n deploy/scripts/deploy_staging.sh bootstrap/host/bootstrap_remote_staging.sh bootstrap/remote/45_configure_github_repo_ci.sh bootstrap/remote/60_deploy_codex_k8s.sh`

## Риски/зависимости
- Зависимости: корректно настроенный GitHub webhook secret.
- Риск: неконсистентность dedup при конкурентной обработке.

## План релиза (верхний уровень)
- Deploy в staging в день реализации, с ручным replay webhook smoke.

## Апрув
- request_id: owner-2026-02-06-day1
- Решение: approved
- Комментарий: Day 1 scope принят.
