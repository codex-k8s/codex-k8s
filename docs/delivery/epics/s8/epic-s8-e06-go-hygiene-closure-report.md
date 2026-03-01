---
doc_id: EPC-CK8S-S8-E06-HYGIENE
type: implementation-report
title: "Sprint S8 E06: Cross-service Go hygiene closure report (Issue #230)"
status: in-review
owner_role: dev
created_at: 2026-02-28
updated_at: 2026-02-28
related_issues: [230, 226, 228]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-28-issue-230-hygiene"
---

# Sprint S8 E06: Cross-service Go hygiene closure report (Issue #230)

## TL;DR
- Выполнен финальный cross-service аудит дублирования по `services/*` и `libs/go/*` после закрытия `#225..#229`.
- Убраны low-risk дубли в `control-plane`/`worker`/`libs/go/registry`, а также в gRPC mapping-слое `control-plane`.
- `tools/lint/dupl-baseline.txt` синхронизирован с актуальным состоянием: объём baseline снижен с `62` до `43` строк.
- Остаточные дубли сведены в приоритизированный debt backlog с owner-decision комментариями.

## Что закрыто в текущем run:dev

### Удалённые low-risk дубли
| ID | Изменение | Файлы | Результат |
|---|---|---|---|
| DUP-230-01 | Вынесен общий helper для разбора image reference | `libs/go/registry/image_ref.go`, `services/internal/control-plane/internal/domain/runtimedeploy/service_build.go`, `services/jobs/worker/internal/app/job_image_registry_checker.go` | Удалён cross-service дублирующийся код `extractRegistryRepositoryPath/splitImageRef`. |
| DUP-230-02 | Вынесен общий helper ожидания Kubernetes job с логированием ошибок | `services/internal/control-plane/internal/domain/runtimedeploy/service_logs.go`, `service_build.go`, `service_repo_sync.go` | Удалён дублирующийся wait+logs error-path в build/repo-sync сценариях. |
| DUP-230-03 | Нормализован gRPC mapping в transport boundary | `services/internal/control-plane/internal/transport/grpc/server_staff_methods.go` | Убраны повторяющиеся маппинги `RepositoryBinding` и `ConfigEntry` через единые helper-caster функции. |
| DUP-230-04 | Нормализована конструкция зависимостей staff-сервиса | `services/internal/control-plane/internal/domain/staff/service.go`, `services/internal/control-plane/internal/app/app.go` | Убрана сигнатурная дубликация аргументов конструктора через `staff.Dependencies`. |

## Сводный self-check (common/go)

### `docs/design-guidelines/common/check_list.md`
- Границы слоёв сохранены: transport/domain/repository разделение не нарушено.
- Helper-код вынесен на корректный уровень переиспользования (`libs/go/registry`, package-level helpers).
- Повторяющиеся блоки сведены в переиспользуемые функции без изменения контрактов.
- Новые внешние зависимости не добавлялись.

### `docs/design-guidelines/go/check_list.md`
- Изменения ограничены структурным рефакторингом без изменения продуктового поведения.
- DTO/transport boundary остались типизированными, без `map[string]any` контрактов.
- Локальные дубли в крупных файлах вынесены в `*_helpers`/service-level helper функции.
- `context.Background()` в изменённые слои не добавлялся.

## Residual technical debt backlog

| Debt ID | Локация | Описание | Приоритет | Owner decision |
|---|---|---|---|---|
| TD-230-01 | `services/external/api-gateway/internal/transport/http/casters/controlplane.go` | Повторяющиеся паттерны ручных DTO-кастеров (несколько dupl-групп). | P1 | Оставить в baseline; выполнить пакетную нормализацию в отдельном bounded issue по transport casters. |
| TD-230-02 | `services/internal/control-plane/internal/domain/staff/service_repository_management.go` | Дублирующиеся циклы загрузки/валидации docset source файлов. | P1 | Вынести в общий docset-file loader helper в отдельном refactor issue. |
| TD-230-03 | `services/internal/control-plane/internal/repository/postgres/{agent,prompttemplate}/repository.go` | Повтор queryOne/CollectRows helper для разных репозиториев. | P1 | Рассмотреть единый generic helper в postgres shared utility слое. |
| TD-230-04 | `services/internal/control-plane/internal/transport/grpc/server_runtime_methods.go` | Повторяющиеся runtime error RPC обвязки. | P2 | Оставить до решения по endpoint deprecation в issue `#81`. |
| TD-230-05 | `services/internal/control-plane/internal/clients/githubmgmt/client.go` | Локальные повторяющиеся request/response блоки. | P2 | Вынести унифицированный request executor в отдельном issue клиента GitHub. |
| TD-230-06 | `services/internal/control-plane/internal/repository/postgres/configentry/repository.go` | Повторяющиеся SQL row mapping-блоки. | P2 | Оставить в baseline, устранить в рамках repo-layer cleanup пакета. |
| TD-230-07 | `services/internal/control-plane/internal/clients/kubernetes/client.go` + `libs/go/k8s/joblauncher/launcher.go` | Частичное дублирование k8s helper-паттернов между сервисом и библиотекой. | P2 | Требуется design decision по ownership helper-слоя (`libs/*` vs service-local). |
| TD-230-08 | `services/internal/control-plane/internal/domain/{mcp,runstatus}/model.go` | Схожие model-фрагменты в соседних bounded contexts. | P2 | Оставить до отдельного domain model alignment, чтобы не смешать контексты. |

## Проверки
- `make dupl-go` — pass (после синхронизации baseline).
- `make lint-go` — pass.
- `go test ./services/internal/control-plane/...` — pass.
- `go test ./services/jobs/worker/...` — pass.
- `go test ./libs/go/registry/...` — pass.

## Owner decisions (предложение)
- OD-230-01: считать `S8-E06` закрытым при текущем baseline, так как low-risk дубли удалены, а остаточные группы классифицированы и приоритизированы.
- OD-230-02: вынести `TD-230-01..TD-230-03` в отдельный follow-up execution stream `run:dev` как следующий шаг hygiene.
