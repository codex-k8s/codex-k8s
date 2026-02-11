---
doc_id: EPC-CK8S-S2-D1
type: epic
title: "Epic S2 Day 1: Migrations, schema ownership and OpenAPI contract-first baseline"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-11
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 1: Migrations, schema ownership and OpenAPI contract-first baseline

## TL;DR
- Цель эпика: привести выполнение миграций и ownership схемы к устойчивой, воспроизводимой модели для многоподовой платформы.
- Ключевая ценность: отсутствуют гонки миграций, worker не стартует на неподготовленной схеме, ownership схемы закреплён за `internal/control-plane`.
- До начала транспортных расширений API/UI: внедрить contract-first OpenAPI baseline (validation + codegen).
- MVP-результат: выбран и задокументирован единый способ миграций для staging (и задел на prod), а также зафиксирован OpenAPI pipeline для backend/frontend.

## Priority
- `P0`.

## Контекст (важное противоречие, требуется решение Owner)
В монорепо миграции должны находиться *внутри держателя схемы*.
Ранее формулировка `cmd/cli/migrations/*.sql` в гайдах была написана в предположении “репозиторий = один сервис”.
Для `codex-k8s` стандарт уточняется:
- миграции лежат в `services/<zone>/<db-owner-service>/cmd/cli/migrations/*.sql`;
- владелец схемы обязан быть один (shared DB без владельца запрещён).

Текущее состояние transport-контрактов:
- `services/external/api-gateway/api/server/api.yaml` существует, но покрывает только webhook ingress;
- runtime валидация по OpenAPI и codegen server/client пока не включены;
- frontend (`services/staff/web-console`) использует ручные API-обёртки.

Решение для S2 Day1:
- до расширения транспорта внешних клиентов внедряем contract-first baseline:
  - backend: `oapi-codegen` + `kin-openapi`,
  - frontend: `@hey-api/openapi-ts` + `@hey-api/client-axios`.
- Выбор библиотек подтверждён по актуальной документации (Context7).

## Решение для MVP
- Владелец схемы: `services/internal/control-plane`.
- Миграции хранятся в `services/internal/control-plane/cmd/cli/migrations/*.sql`.

## Scope
### In scope
- Зафиксировать стандарт размещения миграций в монорепо документально.
- Если остаётся Job: гарантировать идемпотентность и отсутствие параллельного запуска, добавить evidence в runbook.
- Если переходим на initContainer: добавить механизм взаимного исключения (например, advisory lock в Postgres) и гарантию “один мигратор”.
- Для OpenAPI:
  - расширить `api.yaml` до фактически поддерживаемых public/staff endpoint'ов;
  - добавить runtime request/response validation в `api-gateway`;
  - добавить backend codegen (`oapi-codegen`) для DTO/server stubs;
  - добавить frontend codegen (`@hey-api/openapi-ts`) для typed API client;
  - закрепить make/CI шаг регенерации.

### Out of scope
- Production-grade multi-stage migration orchestration (расширенный gate) вне MVP.
- Полный rollout всех planned `run:*` API (кроме текущего S2 scope).

## Критерии приемки эпика
- Миграции выполняются предсказуемо и без гонок.
- Есть явный владелец схемы (control-plane), отражено в доках и в deploy.
- Есть smoke evidence на staging.
- OpenAPI покрывает все активные external/staff endpoint'ы S2 scope.
- Backend и frontend используют сгенерированные артефакты контракта, ручные несогласованные DTO не остаются в транспортном слое.
- В CI есть проверка актуальности codegen-артефактов.
