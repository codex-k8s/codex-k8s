---
doc_id: TRH-CK8S-S8-0001
type: traceability-history
title: "Sprint S8 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [223, 225, 226, 227, 228, 229, 230, 281, 282, 320, 322, 325, 327]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-traceability-s8-history"
---

# Sprint S8 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S8.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #223 (`run:plan`, 2026-02-27)
- Для FR-002/FR-004/FR-033 и NFR-002/NFR-010/NFR-018 добавлен execution-governance пакет Sprint S8 Day1:
  `docs/delivery/epics/s8/epic-s8-day1-go-refactoring-plan.md`,
  `docs/delivery/epics/s8/epic_s8.md`,
  `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md`,
  `docs/delivery/delivery_plan.md`.
- Выполнен plan-аудит Go-кода по сервисам и библиотекам (oversized files, duplicate hotspots, database access alignment).
- Создан параллельный implementation backlog `#225..#230`:
  - `#225` control-plane decomposition;
  - `#226` api-gateway transport cleanup;
  - `#227` worker decomposition;
  - `#228` agent-runner helper normalization;
  - `#229` shared libs pgx/servicescfg alignment;
  - `#230` cross-service hygiene closure.
- Через Context7 (`/websites/cli_github_manual`) подтвержден актуальный CLI-синтаксис `gh issue create`/`gh pr create`/`gh pr edit`; новые внешние зависимости не добавлялись.

## Актуализация по Issue #225 (`run:dev`, 2026-02-28)
- Для FR-002/FR-033 и NFR-002/NFR-010/NFR-018 выполнен рефакторинг bounded scope `S8-E01`:
  декомпозированы oversized-файлы `webhook/service.go`, `staff/service_methods.go`, `transport/grpc/server.go`
  в тематические smaller units без изменения API/proto/OpenAPI контрактов.
- По правилам размещения кода вынесены helper- и methods-блоки в отдельные файлы:
  - `services/internal/control-plane/internal/domain/webhook/service_helpers.go`;
  - `services/internal/control-plane/internal/domain/staff/service_config_entries.go`;
  - `services/internal/control-plane/internal/domain/staff/service_repository_management.go`;
  - `services/internal/control-plane/internal/domain/staff/service_repository_management_types.go`;
  - `services/internal/control-plane/internal/transport/grpc/server_staff_methods.go`;
  - `services/internal/control-plane/internal/transport/grpc/server_runtime_methods.go`.
- Устранён ad-hoc payload в `RunRepositoryPreflight`: локальный `[]struct{...}` заменён на typed модель `servicesYAMLPreflightEnvSlot`.
- Сохранён единый подход к error mapping на транспортной границе:
  gRPC-преобразование ошибок продолжает выполняться через `toStatus`, без локальных межслойных трансляторов в handlers/domain.
- Проверки по изменённому scope:
  - `go test ./services/internal/control-plane/internal/domain/webhook ./services/internal/control-plane/internal/domain/staff ./services/internal/control-plane/internal/transport/grpc`;
  - `go test ./services/internal/control-plane/...`;
  - `make lint-go` (pass);
  - `make dupl-go` (обнаруживает pre-existing дубли в репозитории, включая исторически повторяющиеся блоки вне scope `#225`).
- Через Context7 (`/grpc/grpc-go`) подтверждена актуальная рекомендация по возврату gRPC-ошибок через `status.Error` и сохранению уже типизированных status-кодов.

## Актуализация по Issue #227 (`run:dev`, 2026-02-28)
- Для FR-033 и NFR-018 выполнена декомпозиция worker orchestration-сервиса без изменения продуктового поведения:
  `services/jobs/worker/internal/domain/worker/service.go` разделён на
  `service_queue_cleanup.go`, `service_queue_lifecycle.go`, `service_queue_dispatch.go`, `service_queue_finalize.go`.
- Для сокращения повторов namespace-resolution добавлен package-level helper:
  `services/jobs/worker/internal/domain/worker/service_queue_helpers.go` (`applyPreparedNamespace`),
  и обновлён recovery-путь в `job_not_found_recovery.go`.
- Поведенческое покрытие сохранено: пройдены проверки
  `go test ./services/jobs/worker/internal/domain/worker/...` и `go test ./services/jobs/worker/...`.
- По checklist gate выполнен `make lint-go`.
- `make dupl-go` зафиксировал pre-existing дубли вне scope текущего issue (в `control-plane` и `api-gateway`);
  для изменённого набора файлов `worker` локальная проверка `dupl` не выявила новых дублей.
- Трассируемость синхронизирована с `docs/delivery/issue_map.md` (добавлена строка по Issue `#227`).

## Актуализация по Issue #229 (`run:dev`, 2026-02-28)
- Для FR-004/FR-033 и NFR-018 выполнено выравнивание shared Go-библиотек в bounded scope `S8-E05`:
  `libs/go/postgres` и `libs/go/servicescfg`.
- В `libs/go/postgres` закреплён pgx-native baseline:
  - `OpenPGXPool` остаётся основным API для нового кода;
  - `Open` переведён в explicit compatibility-wrapper `OpenSQLDB` с `Deprecated`-пометкой;
  - добавлен unit coverage (`db_test.go`) для normalization/DSN helper-функций.
- В `libs/go/servicescfg` выполнена модульная декомпозиция без изменения поведения:
  `load.go` разделён на тематические файлы `load_namespace.go`, `load_validation.go`,
  `load_components.go`, `load_context.go`, `load_imports.go`, `load_helpers.go`.
- Релевантный дизайн-гайд обновлён:
  `docs/design-guidelines/go/infrastructure_integration_requirements.md` теперь явно фиксирует правило
  `pgxpool` по умолчанию и `database/sql` только как compatibility-path.
- Проверки по изменённому scope:
  `go test ./libs/go/servicescfg ./libs/go/postgres/...`, `go test ./...`, `make lint-go`.
- `make dupl-go` фиксирует pre-existing дубли вне scope текущего issue (в `control-plane` и `api-gateway`).
- Трассируемость синхронизирована с `docs/delivery/issue_map.md` (добавлена строка по Issue `#229`).

## Актуализация по Issue #230 (`run:dev`, 2026-02-28)
- Для FR-002/FR-004/FR-033 и NFR-002/NFR-010/NFR-018 выполнен финальный consolidating stream `S8-E06`:
  cross-service hygiene closure и residual debt report.
- Удалены low-risk дубли в `control-plane`/`worker`/`libs`:
  - вынесен общий helper `libs/go/registry/image_ref.go` и удалено дублирование `extractRegistryRepositoryPath/splitImageRef`;
  - добавлен общий helper ожидания job с логами ошибок (`waitForJobCompletionWithFailureLogs`) для `runtimedeploy` build/repo-sync path;
  - унифицирован gRPC mapping `RepositoryBinding`/`ConfigEntry` через package-level helper-caster'ы;
  - конструктор `staff.Service` переведён на `staff.Dependencies` для устранения сигнатурной дубликации.
- `tools/lint/dupl-baseline.txt` синхронизирован с текущим кодом:
  baseline сокращён с `62` до `43` строк, удалены устаревшие записи и зафиксированы только актуальные residual duplicates.
- Подготовлен consolidated отчёт:
  `docs/delivery/epics/s8/epic-s8-e06-go-hygiene-closure-report.md`
  (self-check по `common/go` чек-листам, residual debt backlog с приоритетами и owner-decision предложениями).
- Проверки по изменённому scope:
  `make dupl-go`, `make lint-go`, `go test ./services/internal/control-plane/...`,
  `go test ./services/jobs/worker/...`, `go test ./libs/go/registry/...`.
- Трассируемость синхронизирована с `docs/delivery/issue_map.md` (добавлены строки по `#226`, `#228`, `#230`).

## Актуализация по Issue #281 (`run:dev`, planned 2026-03-06)
- Для FR-020/FR-033/FR-049/FR-050 и NFR-010/NFR-018 добавлен execution backlog Sprint S8 Day2:
  `docs/delivery/epics/s8/epic-s8-day2-empty-repository-initialization.md`,
  `docs/delivery/epics/s8/epic_s8.md`,
  `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Зафиксирован отдельный onboarding path для пустого GitHub-репозитория:
  deterministic detection -> bootstrap bundle -> direct initial commit в default branch -> onboarding summary issue.
- В traceability baseline закреплено, что empty-repo onboarding:
  - не использует PR как обязательный init-механизм до появления первой ветки/коммита;
  - обязан создавать typed `services.yaml` и docs scaffold как стартовую точку для stage-flow;
  - обязан быть идемпотентным и сохранять audit trail по созданным файлам и bootstrap SHA.
- Источники требований и решений:
  - FR-020 (per-repo `services.yaml`);
  - FR-049 (repository onboarding preflight);
  - FR-050 (docs tree в prompt context);
  - ADR-0007 (multi-repo composition/docs federation как будущий режим).

## Актуализация по Issue #282 (`run:dev`, planned 2026-03-06)
- Для FR-002/FR-020/FR-033/FR-049/FR-050 и NFR-010/NFR-018 добавлен execution backlog Sprint S8 Day3:
  `docs/delivery/epics/s8/epic-s8-day3-existing-repository-adoption.md`,
  `docs/delivery/epics/s8/epic_s8.md`,
  `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Зафиксирован целевой adoption path для существующего репозитория с кодом:
  repository classification -> deterministic scan report -> dedicated onboarding-task -> PR с `services.yaml` и docs baseline.
- В traceability baseline закреплены ограничения безопасности и качества:
  - onboarding-task имеет ограниченный writable scope и не должна делать произвольный рефакторинг приложения;
  - adoption выполняется только через reviewable PR;
  - rerun должен быть идемпотентным на уровне onboarding branch/PR для одного и того же `ref`.
- Источники требований и решений:
  - FR-002 (provider abstraction);
  - FR-020 (per-repo `services.yaml`);
  - FR-049 (repository onboarding preflight);
  - FR-050 (role-aware docs context);
  - ADR-0007 и S4 Day1 (multi-repo compose/docs federation).

## Актуализация по Issue #320 (`run:plan`, 2026-03-11)
- Для FR-033/FR-049/FR-050 и NFR-010/NFR-018 добавлен plan-package Sprint S8 Day4:
  `docs/delivery/epics/s8/epic-s8-day4-documentation-ia-refactor-plan.md`,
  `docs/delivery/epics/s8/epic_s8.md`,
  `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- План закрепляет единый execution backlog item для documentation governance без полного re-root `docs/`:
  - верхний доменный слой сохраняется как `docs/product|architecture|delivery|ops|templates`;
  - governance source of truth остаётся в `docs/delivery/development_process_requirements.md`;
  - канонический root navigation path = `docs/index.md`, доменные индексы = `docs/<domain>/README.md`.
- В plan baseline зафиксированы обязательные execution-wave результаты:
  - migration-map переносов и affected-links inventory до любого file move;
  - синхронизация `services.yaml/spec.projectDocs` и `spec.roleDocTemplates` после migration;
  - update открытых issues `#254`, `#281`, `#282`, `#309`, `#312`, `#318`, `#322` после смены путей;
  - явная валидация repo-local path refs и stale blob links с evidence в implementation PR.
- Зафиксирована зависимость Sprint S8 onboarding streams от этого plan-package:
  `#281/#282` не должны финализировать docs baseline до merge результата `#320`, иначе будет закреплён неканонический root docs path вместо `docs/index.md`.
- Через Context7 (`/websites/cli_github_manual`) подтверждён актуальный синтаксис `gh issue edit`/`gh pr create`/`gh pr edit` для будущего update open issues и PR-flow; новые внешние зависимости не требуются.

## Актуализация по Issue #320 (`run:dev`, 2026-03-11)
- Для FR-033/FR-049/FR-050 и NFR-010/NFR-018 выполнен implementation package documentation governance:
  `docs/index.md`,
  `docs/{product,architecture,delivery,ops}/README.md`,
  `docs/delivery/documentation_ia_migration_map.md`,
  `docs/architecture/initiatives/**`,
  `docs/ops/handovers/**`,
  `Makefile`,
  `services.yaml`.
- Каноническая IA теперь зафиксирована не только в process requirements, но и в навигационном слое репозитория:
  - root navigation = `docs/index.md`;
  - доменные индексы = `docs/<domain>/README.md`;
  - `docs/templates/index.md` переведён в template-only catalog;
  - устаревшая ссылка в `docs/templates/user_story.md` переписана на `docs/templates/definition_of_done.md`.
- Initiative/stage-specific пакеты разложены по доменным подпапкам:
  - lifecycle agents/prompt templates -> `docs/architecture/initiatives/agents_prompt_templates_lifecycle/*`;
  - Sprint S7 MVP readiness package -> `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/*`;
  - Sprint S6 ops handover -> `docs/ops/handovers/s6/*`.
- Runtime docs context синхронизирован:
  - `services.yaml/spec.projectDocs` получил явную ссылку на `docs/index.md`;
  - внутренние markdown-ссылки и delivery traceability приведены к новым путям;
  - open issue bodies `#281`, `#282`, `#312` очищены от same-repo blob links и branch-specific doc refs.
- Проверки implementation package:
  - `rg -n -g '!docs/delivery/documentation_ia_migration_map.md' -g '!docs/delivery/issue_map.md' -g '!docs/delivery/requirements_traceability.md' 'docs/README\.md|docs/03_engineering/' docs services.yaml` — no matches;
  - `rg -n -g '!docs/delivery/issue_map.md' -g '!docs/delivery/requirements_traceability.md' 'https://github\.com/codex-k8s/kodex/blob/' docs services.yaml` — no matches;
  - `git diff --check` — passed.

## Актуализация по Issue #327 (`run:doc-audit`, 2026-03-12)
- Для FR-033 и NFR-010/NFR-018 выполнена декомпозиция delivery traceability на два уровня:
  current-state master-файлы (`docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`) и sprint-specific history packages в `docs/delivery/traceability/*.md`.
- Добавлен явный индекс `docs/delivery/traceability/README.md` с правилом:
  root-файлы остаются source of truth, а historical packages хранят только evidence/delta и не переопределяют каноническую матрицу.
- Из `docs/delivery/requirements_traceability.md` вынесены historical секции `## Актуализация по Issue ...`;
  корневой файл оставляет только FR/NFR-матрицу, правило актуализации и ссылки на history packages.
- `docs/delivery/issue_map.md` сокращён до master-index без sprint narrative appendices;
  navigation и governance синхронизированы через `docs/index.md`, `docs/delivery/README.md`,
  `docs/delivery/development_process_requirements.md`, `docs/delivery/{sprints,epics}/README.md`.
