---
doc_id: EPC-CK8S-S3-D9
type: epic
title: "Epic S3 Day 9: Declarative full-env deploy and runtime parity"
status: in-progress
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-14
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 9: Declarative full-env deploy and runtime parity

## TL;DR
- Цель: перейти от shell-first деплоя к декларативному `services.yaml` и Go-движку оркестрации full-env.
- Входная обязательная подзадача (из референса): добавить `partials` для шаблонного рендера манифестов без перехода на Helm.
- MVP-результат Day9: детерминированный full-env deploy по typed inventory, reusable шаблоны манифестов, runtime parity для non-prod (`dev/staging/ai-slot`).

## Priority
- `P0`.

## Контекст и проблема
- Сейчас основной контур staging/full-env деплоя завязан на shell-скрипты и ручной рендер `*.yaml.tpl`.
- `services.yaml` в `codex-k8s` пока не является реальным source of truth для состава сервисов и порядка оркестрации.
- В манифестах много повторяющихся блоков (labels/annotations/env/probes/ingress/tls/wait initContainers), которые сложно сопровождать без reusable partial-шаблонов.
- Нужно упростить DX для разработчиков и агентов: меньше копипасты, меньше скрытой логики в bash, больше детерминизма в typed config + Go runtime.

## Scope
### In scope
- D9-T1. Template partials для манифестов (обязательная подзадача из референса):
  - добавить поддержку partial-файлов с `{{ define "..." }}` и вызовами через `{{ template "..." . }}` и helper `include`;
  - подключение partials через `services.yaml` (`templates.partials`) и/или конвенцию (решение фиксируется перед реализацией);
  - сохранить текущий TemplateContext-модель (`.Env`, `.Namespace`, `.Slot`, `.Versions`, ...), не ломая текущие шаблоны;
  - добавить читаемые ошибки при конфликте `define`/ошибках парсинга.
- D9-T2. Typed inventory в `services.yaml` для full-env deploy:
  - описать infra/services/deploy groups/dependencies/overlays как source of truth;
  - закрепить порядок выкладки:
    `stateful dependencies -> migrations -> internal domain services -> edge services -> frontend`.
- D9-T3. Go-orchestrator деплоя:
  - основной путь apply/readiness должен исполняться через Go-движок;
  - shell остается thin-wrapper entrypoint без бизнес-логики оркестрации.
- D9-T4. Runtime parity non-prod:
  - зафиксировать dev/full-env режимы для Go и frontend сервисов;
  - унифицировать поведение окружения agent-run и сервисов слота.
- D9-T5. Документация и трассировка:
  - обновить `docs/architecture/*`, `docs/product/*` (где затронуты process/runtime), `docs/delivery/*`, `docs/ops/staging_runbook.md`;
  - синхронизировать `issue_map` и `requirements_traceability`.

### Out of scope
- Production-оптимизации (autoscaling tuning, cost-optimization).

## Критерии приемки
- Partials:
  - проект может определить один или несколько partial-файлов;
  - partial-шаблоны доступны в любом манифесте при рендере;
  - тесты покрывают: загрузку partials, вызов через `template/include`, конфликт имен.
- Declarative deploy:
  - для `codex-k8s` full-env можно поднять из `services.yaml` без ручной правки deploy shell-скриптов;
  - порядок deploy-этапов детерминирован и подтвержден событиями/логами.
- Runtime parity:
  - non-prod режимы запуска сервисов и agent-run согласованы и воспроизводимы;
  - shared workspace volume и права доступа описаны декларативно.
- Документация:
  - проектная документация и трассировка синхронно обновлены по итогам реализации.

## План реализации (Day9)
1. Ввести модуль рендера шаблонов с partials (без изменения текущего контекста).
2. Добавить typed-модель `services.yaml` для deploy inventory и загрузчик с валидацией.
3. Перенести deploy orchestration в Go-движок с тем же порядком зависимостей.
4. Подключить движок в существующий staging/dev pipeline через thin wrappers.
5. Обновить docs + tests + regression checks.

## DoD (engineering)
- Unit tests:
  - parser/render tests для `services.yaml` и partials;
  - negative cases (conflict/missing template/invalid glob).
- Integration checks:
  - dry-run render/plan;
  - staging deploy path через новый orchestrator.
- Documentation:
  - обновлены связанные продуктовые/архитектурные/delivery/ops документы.

## Референсы и источники
- Внутренние:
  - `services.yaml`
  - `deploy/base/**`
  - `deploy/scripts/deploy_staging.sh`
  - `docs/design-guidelines/common/project_architecture.md`
  - `docs/design-guidelines/go/services_design_requirements.md`
- Внешние референсы (подход, не копирование 1:1):
  - `../codexctl/internal/config/config.go`
  - `../project-example/services.yaml`
  - `../codexctl/internal/engine/*`

## Примечание по реализации
- Подход из `codexctl` используется как референс, но реализация в `codex-k8s` должна быть проще для сопровождения и строго соответствовать текущей архитектуре платформы (control-plane + worker + agent-runner, без возврата к workflow-first парадигме).
