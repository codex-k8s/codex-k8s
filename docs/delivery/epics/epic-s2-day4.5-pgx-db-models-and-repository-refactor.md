---
doc_id: EPC-CK8S-S2-D4.5
type: epic
title: "Epic S2 Day 4.5: PGX repository baseline and db-model rollout"
status: completed
owner_role: EM
created_at: 2026-02-12
updated_at: 2026-02-13
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-13-s2-day4.5"
---

# Epic S2 Day 4.5: PGX repository baseline and db-model rollout

## TL;DR
- Цель: перевести репозиторный слой на единый baseline `pgx` + typed `db`-модели, чтобы убрать хрупкие `Scan/Exec` паттерны и снизить шум в ревью.
- Ограничение: домен не зависит от persistence-моделей; репозиторий не тянет transport-контракты.
- Результат: системный стиль `repository/postgres/*` + `dbmodel` + caster без ad-hoc структур.

## Priority
- `P0` (до масштабирования S2 Day5+).

## Why now
- Кодовая база выросла по количеству SQL-paths и ручных `Scan`.
- Для агентной разработки/ревью нужен детерминированный паттерн, чтобы ИИ не генерировал разношерстный persistence-код.
- Требуется устранить регулярные замечания по `any`/ad-hoc моделям в репозиториях.

## Source of truth
- `docs/design-guidelines/go/services_design_requirements.md`
- `docs/design-guidelines/go/check_list.md`
- `docs/design-guidelines/common/check_list.md`

## Scope
### In scope
- Инвентаризация всех Go-репозиториев с прямым `database/sql`/ручным `Scan`.
- Введение единого шаблона DB-моделей с `db`-тегами в persistence-слое.
- Переход на `pgx`-паттерны там, где это дает упрощение и типобезопасность:
  - `pgxpool` для access слоя;
  - `pgx.CollectRows` + `pgx.RowToStructByName`/`pgx.RowToAddrOfStructByName`;
  - `pgx.NamedArgs` для читаемых write/query операций;
  - batch/tx-paths через `pgx.Batch`/`BeginTx` по необходимости.
- Явные кастеры между `dbmodel` и доменными типами.
- Удаление `any` из репозиторных контрактов и реализации.

### Out of scope
- Переписывание бизнес-логики доменного слоя.
- Изменение product semantics run/label flow.

## Размещение кода (обязательное)
- DB-модели хранить в persistence-пакетах сервиса:
  - `services/<zone>/<service>/internal/repository/postgres/<entity>/dbmodel/*.go`
- SQL остается в `.../sql/*.sql` (как сейчас).
- Маппинг `dbmodel <-> domain` держать в `internal/domain/casters` или `internal/repository/postgres/<entity>/casters` по текущей структуре сервиса.
- Запрещено:
  - объявлять `dbmodel` в `service.go`/`handler.go`;
  - тащить `dbmodel` в transport или domain API.

## Context7 findings (pgx)
По справке `/jackc/pgx` (Context7) для rollout закрепляются практики:
- `pgxpool` как default pool abstraction;
- `CollectRows` + `RowToStructByName` для типизированного чтения в struct с `db`-тегами;
- `NamedArgs` для сложных SQL с множеством параметров;
- отказ от ручных циклов `rows.Next()+Scan` там, где достаточно typed collectors.

## Stories
### Story-1: Repository inventory and target map
- Составить список репозиториев/файлов, где есть ручные scan/nullable any/ad-hoc payload модели.
- Для каждого файла определить target-паттерн (`dbmodel`, caster, pgx API).

### Story-2: Persistence model baseline
- Ввести `dbmodel` для ключевых сущностей control-plane/worker.
- Убрать анонимные/встроенные persistence-структуры из `repository.go`.

### Story-3: PGX API migration
- Перевести read-paths на typed collection (`CollectRows` + `RowToStructByName`).
- Перевести write-paths на типизированные параметры без `any`.
- Для сложных write-sequences использовать `Batch/Tx`.

### Story-4: Guardrails and docs
- Обновить релевантные гайды/checklists, чтобы новый паттерн проверялся до push.
- Добавить примеры "правильно/неправильно" для repository layer.

## Definition of done
- В затронутых репозиториях отсутствуют `any` в persistence-операциях.
- Для новых/переписанных репозиториев введены `dbmodel` с `db`-тегами.
- Проверки `go test ./...`, `make lint-go`, `make dupl-go` проходят.
- Документация/чек-листы обновлены синхронно.

## Acceptance
- Минимум `control-plane` и `worker` покрыты новым паттерном на критических read/write path.
- PR-ревью не содержит замечаний класса "ad-hoc persistence model / any in repository".

## Реализация (2026-02-13)
- В `control-plane` репозиториях `agentrun` и `staffrun` введены typed persistence модели:
  - `internal/repository/postgres/agentrun/dbmodel/*.go` + `casters.go`;
  - `internal/repository/postgres/staffrun/dbmodel/*.go` + `casters.go`.
- Убраны ad-hoc persistence структуры из `repository.go` в пользу `dbmodel` + явных кастеров.
- Добавлены SQL-paths и методы для namespace-cleanup по lifecycle событий:
  - `list_run_ids_by_repository_issue.sql`;
  - `list_run_ids_by_repository_pull_request.sql`.
- Устранены дубли в domain/repository слое (`make dupl-go` зелёный).
- Проверки эпика выполнены:
  - `go test ./...`;
  - `make lint-go`;
  - `make dupl-go`.
