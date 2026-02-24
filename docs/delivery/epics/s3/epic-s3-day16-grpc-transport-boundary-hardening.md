---
doc_id: EPC-CK8S-S3-D16
type: epic
title: "Epic S3 Day 16: gRPC transport boundary hardening (transport -> service -> repository)"
status: completed
owner_role: EM
created_at: 2026-02-17
updated_at: 2026-02-18
related_issues: [45]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S3 Day 16: gRPC transport boundary hardening (transport -> service -> repository)

## TL;DR
- В рамках Issue #45 проведён аудит gRPC-транспорта `control-plane` и зафиксированы нарушения слоя: часть gRPC handlers обращается напрямую к repository.
- Найдено 2 gRPC ручки с прямыми вызовами `user` repository (4 вызова методов репозитория суммарно).
- Цель эпика: убрать прямой доступ transport -> repository, перенести эту логику в доменные сервисы и оставить в gRPC handlers только transport-обязанности (валидация, маппинг, вызов service).

## Priority
- `P0`.

## Нормативная база
- `docs/design-guidelines/go/services_design_requirements.md`
- `docs/design-guidelines/go/grpc.md`
- `docs/design-guidelines/common/design_principles.md`
- `docs/architecture/c4_container.md`

## Найденные нарушения (analysis baseline)

### 1) Прямой доступ gRPC handler к `user` repository
- Файл: `services/internal/control-plane/internal/transport/grpc/server.go`
- Метод gRPC: `ResolveStaffByEmail`
- Прямые вызовы repository:
  - `s.users.GetByEmail(...)`
  - `s.users.UpdateGitHubIdentity(...)`

### 2) Прямой доступ gRPC handler к `user` repository
- Файл: `services/internal/control-plane/internal/transport/grpc/server.go`
- Метод gRPC: `AuthorizeOAuthUser`
- Прямые вызовы repository:
  - `s.users.GetByEmail(...)`
  - `s.users.UpdateGitHubIdentity(...)`

### 3) Смежный architectural smell (не прямой вызов, но нежелательная связность)
- Файл: `services/internal/control-plane/internal/transport/grpc/server.go`
- `transport/grpc` зависит от repo-типов (`userrepo`, `staffrunrepo`, `configentryrepo`) в сигнатурах/кастерах.
- Требуемое направление на рефакторинге: транспорт опирается на domain service interfaces и domain types, без repository-type leakage в transport слой.

## Scope

### In scope
- Ввести/расширить domain service для staff-auth use-case, чтобы gRPC handlers не вызывали `user` repository напрямую.
- Убрать `Users userrepo.Repository` из зависимостей `grpc.Server` и заменить на service-level interface.
- Перенести логику `ResolveStaffByEmail` / `AuthorizeOAuthUser` в domain service с сохранением текущего поведения.
- Снизить связность transport с repository-type моделями:
  - заменить repo-specific типы в transport-сигнатурах на domain/query types;
  - оставить mapping transport <-> domain через явные кастеры.
- Обновить unit-тесты transport/domain на новый dependency graph.

### Out of scope
- Изменение публичных HTTP/gRPC контрактов (`proto`/OpenAPI).
- Функциональные изменения RBAC/аутентификации.
- Массовый рефактор всех transport-пакетов за пределами найденных мест.

## Декомпозиция (Stories/Tasks)
- Story-1: Domain service extraction для staff auth
  - выделить use-case `resolve/authorize oauth user` в `internal/domain/staff` (или отдельный bounded context внутри домена);
  - покрыть unit-тестами service layer.
- Story-2: gRPC transport cleanup
  - оставить в handler только валидацию req, вызов service, mapping response/error.
- Story-3: Repository type leakage cleanup
  - заменить repo-типы в transport-подписи на domain/query types (где это применимо в рамках Day16).
- Story-4: Regression verification
  - прогнать тесты `control-plane` и зафиксировать evidence в PR.

## Критерии приёмки
- В `services/internal/control-plane/internal/transport/grpc/server.go` отсутствуют прямые вызовы repository методов из gRPC handlers.
- `Dependencies` gRPC server не содержит repository интерфейсов для use-cases, которые должны идти через service слой.
- Ошибки по-прежнему маппятся на transport boundary (`toStatus`) без изменения внешнего контракта.
- Поведение `ResolveStaffByEmail` и `AuthorizeOAuthUser` не меняется функционально (регрессионные тесты зелёные).

## Риски и меры
- Риск: незаметная регрессия в OAuth-flow из-за переноса логики.
  - Мера: unit-тесты на service + transport, smoke сценарий на OAuth use-case.
- Риск: частичный рефактор оставит repo-type leakage в transport.
  - Мера: добавить явную проверку в PR (grep по импортам `internal/domain/repository/*` в `transport/grpc`).

## Evidence (Issue #45)
- Поиск импортов repository-пакетов в transport:
  - `rg -n "domain/repository|repository/" services --glob '**/internal/transport/**'`
- Поиск прямых вызовов repository в gRPC server:
  - `rg -n "s\.users\.|userrepo\." services/internal/control-plane/internal/transport/grpc/server.go`

## Реализация (2026-02-18)
- В `services/internal/control-plane/internal/transport/grpc/server.go`:
  - `ResolveStaffByEmail` и `AuthorizeOAuthUser` больше не обращаются к repository напрямую;
  - handlers вызывают domain use-cases `staff.ResolveStaffByEmail(...)` и `staff.AuthorizeOAuthUser(...)`.
- В `services/internal/control-plane/internal/domain/staff/`:
  - добавлен файл `service_staff_auth.go` с use-case логикой OAuth/email resolution;
  - добавлены query-типизированные входы в `services/internal/control-plane/internal/domain/types/query/staff_auth.go`.
- В `grpc.Dependencies` удалена зависимость `Users userrepo.Repository`; транспорт больше не держит `user` repository.
- Дополнительно устранена repository-type leakage по runtime list фильтрам:
  - gRPC transport использует `querytypes.StaffRunListFilter` вместо `staffrunrepo.ListFilter`;
  - сигнатуры и маппинг run-моделей переведены на domain types (`entitytypes.StaffRun`/`entitytypes.User`).

## Acceptance checklist
- [x] В `services/internal/control-plane/internal/transport/grpc/server.go` нет прямых вызовов repository методов в gRPC handlers.
- [x] `Dependencies` gRPC server не содержит repository интерфейсов для OAuth/email resolve use-cases.
- [x] Ошибки продолжают маппиться на transport boundary через `toStatus`.
- [x] Поведение `ResolveStaffByEmail` и `AuthorizeOAuthUser` сохранено (валидация/forbidden/update identity).
