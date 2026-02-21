---
doc_id: ARC-MRM-CK8S-0001
type: architecture-design
title: "codex-k8s — Multi-repo Runtime and Docs Federation Design"
status: proposed
owner_role: SA
created_at: 2026-02-21
updated_at: 2026-02-21
related_issues: [100]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-21-multi-repo-design"
---

# Multi-repo Runtime and Docs Federation Design

## TL;DR
- Для `full-env` нужен один детерминированный `effective services.yaml` на запуск, даже если проект состоит из многих репозиториев.
- Поддерживаются все целевые режимы из Issue #100:
  - монорепо с одним `services.yaml`;
  - multi-repo, где `services.yaml` есть в каждом репозитории;
  - гибрид с центральным repo-компоновщиком и локальными repo-манифестами.
- Размещение документации поддерживается в трех режимах:
  - отдельный docs-репозиторий;
  - docs рядом с сервисами в repo-сервисах;
  - комбинированно.
- Рекомендованный вариант: федеративная компоновка с repo-alias и единым runtime-resolver в `control-plane`.

## Контекст и цель

Issue #100 требует снять текущую зависимость от mono-repo модели и формализовать multi-repo режим для:
- развёртывания сервисов из нескольких репозиториев в dev-slot/full-env;
- поддержки `services.yaml` как в одном репозитории, так и в каждом отдельно;
- поддержки документации в отдельном docs repo, в сервисных repo, и в гибридном варианте.

Цель этого design-пакета:
- зафиксировать единый подход к компоновке runtime-конфигурации и docs-контекста;
- не нарушить текущие архитектурные границы (`external` thin-edge, домен в `internal/control-plane`, reconciliation в `jobs/worker`);
- описать миграцию и runtime-влияние до `run:dev` реализации.

## Нефункциональные ограничения

- Kubernetes-only (FR-001).
- Webhook-driven orchestration (FR-003).
- Единая синхронизация состояния через PostgreSQL (FR-004, FR-013).
- Интеграция с репозиториями через provider interface (FR-002).
- Полный аудит операций и детерминированные policy transitions (FR-032, NFR-010).

## Матрица кейсов

### Кейс A: Монорепо (single repo, single `services.yaml`)

Сценарий:
- проект использует один репозиторий;
- `services.yaml` лежит в корне (или в `services_yaml_path` этого repo).

Ожидаемое поведение:
- resolver использует этот файл как `effective root`;
- docs-контекст собирается из `spec.projectDocs[]` с `repository` по умолчанию = текущий repo;
- текущее поведение платформы сохраняется без изменений.

### Кейс B: Multi-repo (каждый repo имеет свой `services.yaml`)

Сценарий:
- у проекта несколько репозиториев;
- каждый repo описывает свой фрагмент инфраструктуры/сервисов.

Ожидаемое поведение:
- `control-plane` строит `effective root` из repo-фрагментов:
  - либо через явный root-компоновщик;
  - либо через виртуальный root на основе метаданных repositories (если root отсутствует);
- `worker` разворачивает единый execution-plan в namespace запуска.

### Кейс C: Гибрид (центральный orchestrator repo + service repos)

Сценарий:
- есть центральный repo (платформенный или проектный), который описывает общий compose;
- сервисные repo держат локальные `services.yaml` для своих частей.

Ожидаемое поведение:
- central root импортирует service fragments по `repository + path + ref`;
- локальные манифесты остаются source-of-truth для своих сервисов;
- итоговый execution-plan собирается детерминированно в runtime.

### Кейс D: Docs только в отдельном docs repo

Сценарий:
- проектная документация вынесена в выделенный repository;
- service repos содержат код и runtime-манифесты.

Ожидаемое поведение:
- docs-контекст для агентов резолвится из docs repo через `spec.projectDocs[]`;
- runtime-deploy не зависит от расположения docs.

### Кейс E: Docs только в сервисных repo

Сценарий:
- документация хранится рядом с сервисами в каждом repo.

Ожидаемое поведение:
- `spec.projectDocs[]` содержит repo-aware пути;
- агент получает role-aware подборку из нескольких repo.

### Кейс F: Комбинированный docs режим

Сценарий:
- часть документов централизована в docs repo (policy/standards);
- часть живет в service repos (service-specific design/ops).

Ожидаемое поведение:
- resolver собирает docs по приоритету;
- конфликты разрешаются предсказуемо (policy-first, затем service-specific overrides).

## Рекомендуемый вариант (best option)

Рекомендуется вариант **federated composition**:
- один `effective services.yaml` на запуск;
- источники манифеста и docs могут быть распределены по разным репозиториям;
- компоновка выполняется в домене `control-plane` с idempotent execution в `worker`.

Почему этот вариант:
- покрывает все кейсы A..F без отдельного режима выполнения;
- не ломает FR-022 (сам `codex-k8s` может оставаться монорепо);
- сохраняет FR-020 (per-repo `services.yaml`) и добавляет управляемый кросс-repo deploy;
- удерживает архитектурные границы и auditability.

## Контракт `services.yaml`: расширение для federation

Предлагаемое расширение (концептуально):

```yaml
spec:
  repositories:
    - alias: orchestrator
      provider: github
      owner: org
      name: project-control
      role: orchestrator
      defaultRef: main
      servicesPath: services.yaml
    - alias: billing
      provider: github
      owner: org
      name: billing-service
      role: service
      defaultRef: main
      servicesPath: deploy/services.yaml
    - alias: docs
      provider: github
      owner: org
      name: project-docs
      role: docs
      defaultRef: main

  composition:
    rootRepository: orchestrator
    imports:
      - repository: billing
        path: deploy/services.yaml
        ref: main
        required: true

  projectDocs:
    - repository: docs
      path: docs/architecture
      description: "Базовая архитектурная документация проекта"
      roles: [sa, em, dev, qa]
      optional: false
    - repository: billing
      path: docs
      description: "Сервисная документация billing"
      roles: [sa, dev, qa, sre]
      optional: true
```

Правила:
- Если `composition.rootRepository` не задан, root строится виртуально из всех `role=service`.
- `repository` в `projectDocs[]` обязателен в multi-repo режиме.
- Для монорепо поле `repository` опционально (по умолчанию текущий repo).

## Алгоритм runtime-resolve (full-env/dev-slot)

1. Получить связанный набор repositories проекта из БД.
2. Определить entrypoint:
   - explicit `rootRepository`, либо
   - virtual root (service repos ordered by deterministic key).
3. Загрузить `services.yaml` root и все `imports` (repo/path/ref).
4. Построить `effective manifest` (typed merge + cycle detection + strict validation).
5. Сформировать execution-plan (deploy order и зависимости).
6. Сформировать repo checkout plan для worker/agent-runner:
   - только нужные repo/path;
   - pin по commit SHA после резолва `ref`.
7. Применить reconciliation в namespace.
8. Записать audit события компоновки и deploy.

## Алгоритм docs-context resolve (role-aware)

1. Взять `spec.projectDocs[]` из `effective manifest`.
2. Отфильтровать по `roles[]` и task context.
3. Сформировать единый список docs refs из нескольких repo.
4. Применить дедупликацию и приоритет:
   - policy/docs repo -> orchestrator repo -> service repos.
5. Передать в prompt context ограниченный список ссылок (size guard).

## Границы сервисов (без изменений архитектурных зон)

`services/external/api-gateway`:
- только вход, валидация, authn/authz, маршрутизация;
- без логики multi-repo resolve.

`services/internal/control-plane`:
- резолв topology/composition;
- формирование `effective manifest` и execution-plan;
- policy/audit decisions.

`services/jobs/worker`:
- checkout/sync нескольких repo по execution-plan;
- idempotent reconcile в namespace;
- ретраи/lease/cleanup.

## Изменения data model (проектирование)

Расширение `repositories`:
- `alias` (уникально в проекте, стабильный ключ для imports/docs refs),
- `role` (`orchestrator|service|docs|mixed`),
- `default_ref` (ветка/тег по умолчанию),
- `docs_root_path` (опциональный дефолтный root docs).

Новая сущность `repository_compositions` (или эквивалентный JSONB-реестр):
- фиксирует resolved graph (`root`, `imports`, `resolved_commits`),
- хранит последнюю валидированную компоновку для быстрых preflight/runtime checks.

Новая сущность `repository_doc_sources`:
- role-aware docs roots с привязкой к repo alias;
- используется для явного governance docs-контекста.

## Изменения API/контрактов (проектирование)

Staff API additions (private):
- расширение CRUD repositories: поля `alias`, `role`, `default_ref`, `docs_root_path`;
- `POST /api/v1/staff/projects/{project_id}/composition/preview`:
  dry-run resolve `effective manifest`, graph validation, conflict report;
- `GET /api/v1/staff/projects/{project_id}/docs/sources`:
  просмотр эффективного docs graph для роли.

## Security и policy

- Все cross-repo чтения ограничены repository bindings проекта.
- `imports` разрешены только на зарегистрированные repo aliases.
- Для deploy используется pin на commit SHA после первого resolve.
- Запрещены path traversal и выход за разрешённые repo roots.
- Все resolve/deploy действия аудитируются (`flow_events` + links).

## Observability

Новые события (предложение):
- `run.composition.resolve_started`
- `run.composition.resolve_succeeded`
- `run.composition.resolve_failed`
- `run.docs_graph.resolved`
- `run.repo.checkout.plan_built`

Метрики:
- composition resolve latency;
- number of imported repos per run;
- docs refs count before/after dedup;
- deploy failures by repo alias.

## Миграция и rollout

Фаза 1:
- DB schema и staff DTO для repo topology полей.

Фаза 2:
- control-plane resolver + composition preview API.

Фаза 3:
- worker multi-repo checkout/reconcile.

Фаза 4:
- prompt/docs context federation в agent-runner.

Стратегия включения:
- feature flag на проект (`multi_repo_mode=off|preview|enforced`);
- сначала preview + audit-only, затем enforced.

## Runtime impact

- Увеличивается число GitHub API вызовов и операций checkout.
- Появляется зависимость от корректности refs/imports между repo.
- Снижается риск drift за счёт единого `effective manifest` и явной трассировки.

## Риски и компромиссы

Риск:
- рост сложности конфигурации и диагностики.

Компромисс:
- добавляем больше metadata в repositories и отдельный composition preview path.

Mitigation:
- deterministic resolver;
- preflight validation до запуска deploy;
- clear error taxonomy (`failed_precondition`, `conflict`, `not_found`).

## Внешняя валидация решений (Context7)

Проверено по актуальной документации:
- `github.com/google/go-github/v82`:
  - чтение файлов по `path + ref` (`Repositories.GetContents`);
  - обход дерева файлов (`Git.GetTree`, recursive);
  - обработка rate limiting (`RateLimitError`, `AbuseRateLimitError`), conditional requests.
- `k8s.io/client-go`:
  - рекомендованный controller pattern (informer + rate-limited workqueue);
  - retry/conflict handling и `QPS/Burst` ограничения в `rest.Config`.

Эти механизмы достаточны для реализации resolver/reconcile без новых обязательных внешних зависимостей.

## Acceptance criteria для будущего `run:dev`

1. Проект с 3+ repo разворачивается в один dev-slot из `effective manifest`.
2. Monorepo проект продолжает работать без дополнительных полей.
3. Repo-per-service режим работает без central root за счёт virtual root compose.
4. Docs context собирается из docs repo + service repos по role-aware policy.
5. Все переходы аудируются и отображаются в staff runtime diagnostics.

## Связанные документы

- `docs/architecture/adr/ADR-0007-multi-repo-composition-and-docs-federation.md`
- `docs/architecture/data_model.md`
- `docs/architecture/api_contract.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/product/requirements_machine_driven.md` (FR-020, FR-022)
