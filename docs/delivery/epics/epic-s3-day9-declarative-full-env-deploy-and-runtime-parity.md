---
doc_id: EPC-CK8S-S3-D9
type: epic
title: "Epic S3 Day 9: Declarative full-env deploy and runtime parity"
status: completed
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
- Цель: перейти от shell-first деплоя/установки к декларативному `services.yaml` и Go-движку оркестрации full-env.
- Входная обязательная подзадача: добавить `partials` для шаблонного рендера без Helm.
- Day9 охватывает не только deploy, но и bootstrap-first-install на новой машине, чтобы подход был применим для любых проектов на `codex-k8s`.
- Для bootstrap самого `codex-k8s` вводится отдельный консольный install-бинарник с интерактивным дозапросом недостающих переменных/секретов.
- MVP-результат Day9: детерминированные bootstrap + deploy + runtime parity для non-prod (`dev/staging/ai-slot`) на основе typed inventory.

## Priority
- `P0`.

## Контекст и проблема
- Сейчас основной контур staging/full-env деплоя завязан на shell-скрипты и ручной рендер `*.yaml.tpl`.
- Bootstrap нового окружения (`bootstrap/remote/**.sh`) также реализован shell-first и слабо переиспользуем между проектами.
- `services.yaml` в `codex-k8s` пока не является реальным source of truth для состава сервисов и порядка оркестрации.
- В манифестах много повторяющихся блоков (labels/annotations/env/probes/ingress/tls/wait initContainers), которые сложно сопровождать без reusable partial-шаблонов.
- Нужно упростить DX для разработчиков и агентов: меньше копипасты, меньше скрытой логики в bash, больше детерминизма в typed config + Go runtime на всем цикле `bootstrap -> deploy -> AI-slot development`.

## Scope
### In scope
- D9-T1. Template partials для манифестов (обязательная подзадача из референса):
  - добавить поддержку partial-файлов с `{{ define "..." }}` и вызовами через `{{ template "..." . }}` и helper `include`;
  - подключение partials только через явный конфиг в `services.yaml`: `templates.partials`;
  - partials доступны для всех template-consumers: deploy manifests, prompt templates, codex config templates, hook templates;
  - сохранить текущий TemplateContext-модель (`.Env`, `.Namespace`, `.Slot`, `.Versions`, ...), не ломая текущие шаблоны;
  - добавить fail-fast ошибки при конфликте `define`/ошибках парсинга.
- D9-T2. Typed inventory в `services.yaml` для full-env deploy:
  - описать infra/services/bootstrap/deploy groups/dependencies/overlays как source of truth;
  - закрепить порядок выкладки:
    `stateful dependencies -> migrations -> internal domain services -> edge services -> frontend`.
- D9-T3. Go-orchestrator установки и деплоя:
  - основной путь bootstrap/apply/readiness исполняется через Go-движок;
  - shell остается thin-wrapper entrypoint без бизнес-логики оркестрации;
  - логика из `deploy/scripts/**` и `bootstrap/remote/**` переносится в декларативный контур в рамках Day9 (одним проходом, без отложенной "второй волны").
- D9-T3.1 Bootstrap CLI binary (install UX):
  - добавить install CLI-бинарник для первичной установки `codex-k8s` и последующей переинициализации окружения;
  - CLI реализуется на `github.com/spf13/cobra v1.10.2`;
  - CLI запускается локально с машины разработчика/оператора, подключается по SSH к чистому Ubuntu серверу и оркестрирует шаги `bootstrap/remote/*.sh`;
  - CLI поддерживает консольный дозапрос недостающих обязательных параметров (`CODEXK8S_*`, токены, секреты, домен, SSH/K8s контекст и т.д.);
  - CLI умеет читать существующий `bootstrap/host/config.env`, валидировать заполненность и интерактивно дополнять отсутствующие значения;
  - зафиксированный набор команд CLI: `install`, `validate`, `reconcile`;
  - результатом CLI является детерминированный запуск bootstrap/deploy orchestration без ручного редактирования shell-скриптов;
  - runtime full-env deploy для задач (issue/pr slot lifecycle) выполняется платформой (`control-plane + worker`), а не через ручной запуск bootstrap CLI.
- D9-T4. Runtime parity non-prod:
  - для всех non-prod окружений (`dev`, `staging`, `ai-slot`) обязателен hot-reload для Go и frontend сервисов;
  - `prod` остается без hot-reload;
  - унифицировать поведение окружения agent-run и сервисов слота для live-debug.
- D9-T5. Размещение нового модуля конфигурации и рендера:
  - реализовать движок в `libs/go/servicescfg` как переиспользуемый слой для worker/bootstrap/control-plane/CLI-контуров.
- D9-T6. Документация и трассировка:
  - обновить `docs/architecture/*`, `docs/product/*` (где затронуты process/runtime), `docs/delivery/*`, `docs/ops/staging_runbook.md`;
  - синхронизировать `issue_map` и `requirements_traceability`.

### Out of scope
- Production-оптимизации (autoscaling tuning, cost-optimization).

## Критерии приемки
- Partials:
  - проект может определить один или несколько partial-файлов через `templates.partials`;
  - partial-шаблоны доступны во всех template-consumers;
  - тесты покрывают: загрузку partials, вызов через `template/include`, fail-fast при конфликте имен.
- Declarative bootstrap/deploy:
  - для `codex-k8s` и нового проекта полный цикл `bootstrap-first-install + full-env deploy` поднимается из `services.yaml` без ручной правки shell-скриптов;
  - порядок bootstrap/deploy этапов детерминирован и подтвержден событиями/логами.
- Bootstrap CLI:
  - доступен отдельный install-бинарник с командами `install`, `validate`, `reconcile`;
  - при отсутствии обязательных данных CLI задает интерактивные вопросы в консоли и сохраняет значения в конфиг;
  - CLI использует `cobra v1.10.2` и формирует единый UX для локального оператора и агента.
- Runtime parity:
  - во всех non-prod окружениях включен hot-reload для Go и frontend сервисов;
  - non-prod режимы запуска сервисов и agent-run согласованы и воспроизводимы;
  - shared workspace volume и права доступа описаны декларативно.
- Документация:
  - проектная документация и трассировка синхронно обновлены по итогам реализации.

## План реализации (Day9)
1. Ввести модуль `libs/go/servicescfg` с рендером partials, `include` helper и fail-fast валидаторами.
2. Добавить typed-модель `services.yaml` для `bootstrap + deploy + runtime parity` inventory и загрузчик с валидацией.
3. Реализовать install CLI на `cobra v1.10.2` с интерактивным config/secrets intake.
4. Перенести в Go-orchestrator всю бизнес-логику из `deploy/scripts/**` и `bootstrap/remote/**` (одним проходом).
5. Подключить thin-wrapper shell entrypoints к новому движку без дублирования оркестрации.
6. Зафиксировать hot-reload для всех non-prod окружений.
7. Обновить docs + tests + regression checks.

## DoD (engineering)
- Unit tests:
  - parser/render tests для `services.yaml` и partials;
  - negative cases (conflict/missing template/invalid glob).
- Integration checks:
  - dry-run render/plan;
  - интерактивный bootstrap CLI flow с неполным конфигом;
  - bootstrap-first-install path через новый orchestrator;
  - staging deploy path через новый orchestrator;
  - smoke сценарий разработки в AI-slot с hot-reload.
- Documentation:
  - обновлены связанные продуктовые/архитектурные/delivery/ops документы.

## Зафиксированные решения перед реализацией
- `templates.partials`: только явный список в `services.yaml`.
- Partials активны для всех template-consumers.
- Поддерживается `template` + `include` helper.
- Конфликты `define` обрабатываются только fail-fast.
- Общий движок размещается в `libs/go/servicescfg`.
- Day9 выполняется полным проходом, без инкрементального переноса "на потом".
- Hot-reload обязателен для всех non-prod окружений, включая staging; `prod` без hot-reload.

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

## Статус выполнения
- `D9-T1` выполнен:
  - реализован `libs/go/servicescfg` (`load`, `render`, `partials`, `include`, conflict fail-fast);
  - добавлены unit-тесты на partials/load/conflict.
- `D9-T2` выполнен:
  - `services.yaml` переведен на typed inventory (`templates`, `bootstrap`, `deploy`, `runtime`).
- `D9-T3` выполнен:
  - внедрён `bin/codex-bootstrap` с командами `install`, `validate`, `reconcile`;
  - `deploy/scripts/deploy_staging.sh` переведен на thin-wrapper и вызывает только `codex-bootstrap reconcile`.
- `D9-T3.1` выполнен:
  - bootstrap remote path использует declarative reconcile;
  - зафиксировано разделение ответственности: `codex-bootstrap` используется для first-install/операторского reconcile с локальной машины, а runtime-разворачивание task namespace остаётся в контуре платформы (`control-plane + worker`);
  - `bootstrap/remote/00_prepare_host.sh` устанавливает Go toolchain (`CODEXK8S_GO_VERSION`, default `1.25.6`) для запуска CLI на чистом хосте.
- `D9-T4` выполнен:
  - runtime parity зафиксирован в `services.yaml` (`runtime.non_prod`/`runtime.prod`).
- `D9-T5` выполнен:
  - общий движок размещён в `libs/go/servicescfg`.
- `D9-T6` выполнен:
  - обновлены `README.md`, `bootstrap/README.md`, `docs/ops/staging_runbook.md`, `docs/design-guidelines/common/external_dependencies_catalog.md`;
  - прогресс Sprint S3/каталог эпиков и map-трассировка синхронизированы.
