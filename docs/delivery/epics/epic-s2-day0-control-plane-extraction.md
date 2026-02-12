---
doc_id: EPC-CK8S-S2-D0
type: epic
title: "Epic S2 Day 0: Control-plane extraction and thin-edge api-gateway"
status: completed
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-11
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-11-s2-day0"
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
- Ввести gRPC контракт внутреннего sync API в `proto/` (single source of truth) и сгенерировать Go-код в `proto/gen/go/**`.
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
- Story-1: proto контракт для control-plane (webhook ingest, staff APIs, auth bridge).
- Story-2: перенос доменных сервисов и репозиториев из `api-gateway` в `control-plane`.
- Story-3: gRPC server в `control-plane` и gRPC client в `api-gateway`.
- Story-4: deploy wiring: сборка бинарника control-plane и деплой 2х сервисов в staging.
- Story-5: документация: evidence/verification + каталог внешних зависимостей.

## Data model impact (по шаблону data_model.md)
- Схема БД: без изменений.
- Ownership: формально закрепить владельца схемы за `internal/control-plane` (в коде и в доках).
- Миграции: расположение миграций соответствует модели “держатель схемы внутри сервиса”:
  `services/internal/control-plane/cmd/cli/migrations/*.sql`.

## Критерии приемки эпика
- В `services/external/api-gateway` отсутствует доменная логика (use-cases) и прямые реализации postgres-репозиториев.
- `services/internal/control-plane` реализует доменные use-cases, подключается к PostgreSQL и обслуживает gRPC API.
- `api-gateway`:
  - валидирует подпись GitHub webhook;
  - проксирует webhook ingest и staff APIs в `control-plane` по gRPC;
  - выпускает JWT (OAuth callback) и для allowlist/обновления GitHub identity обращается в `control-plane`.
- Миграции схемы лежат внутри держателя схемы и применяются через `goose` из образа.
- `go test ./...` зелёный.

## Риски/зависимости
- Риск: рост объёма изменения (перенос пакетов) может затронуть CI/deploy.
- Зависимость: требуется чёткий proto контракт; без него перенос превратится в ad-hoc вызовы.

## Evidence
- Proto контракт:
  - `proto/codexk8s/controlplane/v1/controlplane.proto`
  - `proto/gen/go/codexk8s/controlplane/v1/controlplane.pb.go`
  - `proto/gen/go/codexk8s/controlplane/v1/controlplane_grpc.pb.go`
- Control-plane (DB owner + домен + gRPC):
  - `services/internal/control-plane/cmd/control-plane/main.go`
  - `services/internal/control-plane/internal/app/app.go`
  - `services/internal/control-plane/internal/app/config.go`
  - `services/internal/control-plane/internal/transport/grpc/server.go`
  - домен и repo impl: `services/internal/control-plane/internal/domain/**`, `services/internal/control-plane/internal/repository/postgres/**`
- API-gateway (thin-edge + gRPC client):
  - `services/external/api-gateway/internal/controlplane/client.go`
  - `services/external/api-gateway/internal/app/app.go`
  - `services/external/api-gateway/internal/app/config.go`
  - staff auth (OAuth+JWT): `services/external/api-gateway/internal/auth/service.go`
  - `services/external/api-gateway/internal/transport/http/server.go`
  - `services/external/api-gateway/internal/transport/http/webhook_handler.go`
  - `services/external/api-gateway/internal/transport/http/staff_handler.go`
- Миграции держателя схемы:
  - `services/internal/control-plane/cmd/cli/migrations/*.sql`
  - `deploy/scripts/deploy_staging.sh` создаёт `configmap/codex-k8s-migrations` из этого пути
  - `deploy/base/codex-k8s/migrate-job.yaml.tpl` применяет миграции через `goose -dir /migrations up`
- Сборка/деплой:
  - отдельные Dockerfile по сервисам:
    - `services/external/api-gateway/Dockerfile`
    - `services/internal/control-plane/Dockerfile`
    - `services/jobs/worker/Dockerfile`
  - `deploy/base/codex-k8s/app.yaml.tpl` и `deploy/base/codex-k8s/migrate-job.yaml.tpl` используют отдельные образа
    (`CODEXK8S_API_GATEWAY_IMAGE`, `CODEXK8S_CONTROL_PLANE_IMAGE`, `CODEXK8S_WORKER_IMAGE`)
- NetworkPolicy baseline:
  - `deploy/base/network-policies/platform-baseline.yaml.tpl` разрешает egress `api-gateway` -> `control-plane` (gRPC 9090, HTTP 8081)
- Каталог внешних зависимостей:
  - `docs/design-guidelines/common/external_dependencies_catalog.md` (grpc/protobuf)

## Verification
- Unit tests: `go test ./...`
- Bash syntax: `bash -n deploy/scripts/deploy_staging.sh`
- Staging smoke: ручной smoke/regression по runbook (OK после деплоя)

## План релиза (верхний уровень)
- Контур dev/staging до dogfooding: см. `.local/agents-temp-dev-rules.md`.
- После merge: push в `codex/dev` должен приводить к автоматическому deploy на staging.
- Smoke: webhook ingest + staff UI базовые сценарии (проверка логов `codex-k8s` и `codex-k8s-control-plane`).

## Апрув
- request_id: owner-2026-02-11-s2-day0
- Решение: approved
- Комментарий: Day 0 scope закрыт, архитектурное выравнивание и перенос ownership схемы приняты.
