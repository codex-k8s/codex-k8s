---
doc_id: EPC-CK8S-S2-D0
type: epic
title: "Epic S2 Day 0: Control-plane extraction and thin-edge api-gateway"
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

# Epic S2 Day 0: Control-plane extraction and thin-edge api-gateway

## TL;DR
- Цель эпика: вернуть архитектуру к стандарту `docs/design-guidelines/common/project_architecture.md`.
- Ключевая ценность: `external/api-gateway` становится thin-edge, а домен/репозитории/БД ownership переезжают в `internal/control-plane`.
- MVP-результат: `api-gateway` валидирует вход и маршрутизирует запросы в `control-plane` по gRPC; `control-plane` владеет доменной логикой и Postgres.

## Priority
- `P0`.

## Контекст
- Фактическое отклонение: доменная логика и репозитории сейчас находятся в `services/external/api-gateway/internal/**`, а `services/internal/control-plane` является placeholder.
- Требование: thin-edge для `external|staff` (валидация/auth/маршрутизация, без orchestration в transport слое).
- Сопутствующее выравнивание: миграции схемы (goose) перенесены в держателя схемы
  `services/internal/control-plane/cmd/cli/migrations/*.sql` (вместо корня репозитория), а staging deploy берёт миграции из этого пути.

## Scope
### In scope
- Ввести gRPC контракт внутреннего sync API в `proto/` (single source of truth).
- Реализовать `services/internal/control-plane` как сервис:
  - `internal/domain/**` (use-cases/ports),
  - `internal/repository/postgres/**` (repo impl),
  - `internal/transport/grpc/**` (server),
  - wiring в `internal/app/**`.
- Перевести `services/external/api-gateway` на модель thin-edge:
  - оставить HTTP transport/middleware/валидацию подписи webhook;
  - заменить прямые вызовы домена/репозиториев на gRPC client в `control-plane`.
- Обновить C4 container и/или API contract, если меняется взаимодействие сервисов.

### Out of scope
- Полная реализация всех bounded contexts из `services_design_requirements.md` (в Day 0 достаточно вынести то, что уже есть).

## Декомпозиция (Stories/Tasks)
- Story-1: proto контракт для control-plane (минимум: webhook ingest, staff read APIs, auth bridge где нужно).
- Story-2: перенос доменных сервисов и репозиториев из `api-gateway` в `control-plane`.
- Story-3: gRPC server в `control-plane` и gRPC client в `api-gateway`.
- Story-4: staging deploy smoke: webhook ingest и staff UI продолжают работать.

## Data model impact (по шаблону data_model.md)
- Схема БД: без изменений.
- Ownership: формально закрепить владельца схемы за `internal/control-plane` (в коде и в доках).
- Миграции: расположение миграций соответствует модели “держатель схемы внутри сервиса”:
  `services/internal/control-plane/cmd/cli/migrations/*.sql`.

## Критерии приемки эпика
- В `services/external/api-gateway` отсутствует доменная логика (use-cases) и прямые реализации postgres-репозиториев.
- `services/internal/control-plane` обрабатывает реальные запросы и подключается к PostgreSQL.
- Staging smoke по базовым сценариям проходит.

## Риски/зависимости
- Риск: рост объёма изменения (перенос пакетов) может затронуть CI/deploy.
- Зависимость: требуется чёткий proto контракт; без него перенос превратится в ad-hoc вызовы.
