---
doc_id: EPC-CK8S-S2-D1
type: epic
title: "Epic S2 Day 1: Migrations, schema ownership and deployment strategy"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-10
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 1: Migrations, schema ownership and deployment strategy

## TL;DR
- Цель эпика: привести выполнение миграций и ownership схемы к устойчивой, воспроизводимой модели для многоподовой платформы.
- Ключевая ценность: отсутствуют гонки миграций, worker не стартует на неподготовленной схеме, ownership схемы закреплён за `internal/control-plane`.
- MVP-результат: выбран и задокументирован единый способ миграций для staging (и задел на prod).

## Priority
- `P0`.

## Контекст (важное противоречие, требуется решение Owner)
В монорепо миграции должны находиться *внутри держателя схемы*.
Ранее формулировка `cmd/cli/migrations/*.sql` в гайдах была написана в предположении “репозиторий = один сервис”.
Для `codex-k8s` стандарт уточняется:
- миграции лежат в `services/<zone>/<db-owner-service>/cmd/cli/migrations/*.sql`;
- владелец схемы обязан быть один (shared DB без владельца запрещён).

## Решение для MVP
- Владелец схемы: `services/internal/control-plane`.
- Миграции хранятся в `services/internal/control-plane/cmd/cli/migrations/*.sql`.

## Scope
### In scope
- Зафиксировать стандарт размещения миграций в монорепо документально.
- Если остаётся Job: гарантировать идемпотентность и отсутствие параллельного запуска, добавить evidence в runbook.
- Если переходим на initContainer: добавить механизм взаимного исключения (например, advisory lock в Postgres) и гарантию “один мигратор”.

### Out of scope
- Production-grade multi-stage migration orchestration (расширенный gate) вне MVP.

## Критерии приемки эпика
- Миграции выполняются предсказуемо и без гонок.
- Есть явный владелец схемы (control-plane), отражено в доках и в deploy.
- Есть smoke evidence на staging.
