---
doc_id: EPC-CK8S-S3-D4
type: epic
title: "Epic S3 Day 4: MCP database lifecycle (create/delete/describe per env)"
status: completed
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 4: MCP database lifecycle (create/delete/describe per env)

## TL;DR
- Цель: дать управляемый инструмент для создания и удаления БД в выбранном окружении.
- MVP-результат: DB lifecycle операции стандартизованы, аудируются и защищены апрувом.

## Priority
- `P0`.

## Scope
### In scope
- MCP tool `database.lifecycle` (`create`, `delete`, `describe`) с environment scoping.
- Policy checks (allowed envs, naming rules, ownership checks).
- Safeguards для destructive операций (`delete` только с явным подтверждением).
- Аудит и traceability в `flow_events`/`links`.

### Out of scope
- Полноценный DBaaS и автоматический backup/restore orchestration.

## Критерии приемки
- Создание/удаление БД воспроизводимы и проходят через approval flow.
- Операции отражаются в UI и audit с `correlation_id`.

## Фактический результат (выполнено)
- MCP tool приведён к каноническому имени:
  - `database.lifecycle`.
- Расширен контракт входа/выхода:
  - actions: `create`, `delete`, `describe`;
  - для `delete` обязателен `confirm_delete=true`;
  - в ответ добавлены `exists`, `owned_by_project`, `owner_project_id` для диагностируемого `describe`.
- Внедрены policy checks до постановки approval:
  - allowlist окружений (`CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS`, fallback: `dev,staging,ai-staging,prod`);
  - валидация `database_name` по regex и длине;
  - ownership checks по проекту/окружению.
- Добавлен ownership-контур в БД и repository-слой:
  - новая таблица `project_databases` (миграция `20260213210000_day13_project_databases.sql`);
  - PostgreSQL repository `projectdatabase` + доменные типы/интерфейсы.
- Реализована run-safe apply логика:
  - `create`: idempotent `EnsureDatabase` + upsert ownership;
  - `delete`: только при `confirm_delete`, с проверкой ownership и cleanup ownership row;
  - `describe`: read-only путь без approval side effects.
- Усилена валидация approve/apply шага:
  - проверка `project_id` в payload против run context при применении approved action.
- Runtime/deploy wiring:
  - новый env `CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS` добавлен в bootstrap/deploy/workflows и в runtime secret прокидку.

## Проверки
- `go test ./services/internal/control-plane/internal/domain/mcp/...` — passed.
- `go test ./services/internal/control-plane/internal/clients/postgresadmin/...` — passed.
- `go test ./services/internal/control-plane/...` — passed.
- `make lint-go` — passed.
- `make dupl-go` — passed.
