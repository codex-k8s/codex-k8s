# Инструкции для ИИ-агентов (обязательно)

## Назначение

Этот файл задает обязательные правила работы с репозиторием `codex-k8s`.
Цель: вести код и документацию к целевой архитектуре cloud-сервиса,
который оркестрирует агентные процессы в Kubernetes и заменяет связку
`codexctl` + часть сценариев `github.com/codex-k8s/yaml-mcp-server`.

## Главные правила

- Перед изменениями читать `docs/design-guidelines/AGENTS.md`.
- Временные правила текущего ручного dev/staging цикла (до полного dogfooding через `run:dev`) см. `.local/agents-temp-dev-rules.md`. Править `.local/agents-temp-dev-rules.md` строго запрещено, если не стоит явная задача на изменение временных правил.
- Для Go-изменений обязательно читать профильные документы из `docs/design-guidelines/go/`.
- Для frontend-изменений обязательно читать `docs/design-guidelines/vue/` и `docs/design-guidelines/visual/`.
- Для инфраструктурных изменений читать `docs/design-guidelines/common/`.
- Для выбора/обновления внешних библиотек читать `docs/design-guidelines/common/external_dependencies_catalog.md`.
- Для планирования и ведения спринта/документации читать `docs/delivery/development_process_requirements.md`.
- Если запрос пользователя противоречит гайдам, приостановить правки и предложить варианты решения.
- Если контекст сессии был сжат/потерян (например, `context compacted`) или есть сомнение, что требования/архитектура актуальны:
  - перечитать `AGENTS.md` и `docs/design-guidelines/AGENTS.md`;
  - перечитать релевантные гайды по области изменения (`docs/design-guidelines/{go,vue,visual,common}/`);
  - сверить задачу с `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`;
  - только после этого планировать и править код.
- Не редактировать сами гайды без явной задачи на изменение стандартов.
- При разработке и доработке проектной документации (бизнес-документов), сверять ее с `docs/research/src_idea-machine_driven_company_requirements.md`. `docs/research/src_idea-machine_driven_company_requirements.md` - это документ, перенесенный из изначального репозитория
`github.com/codex-k8s/codexctl` (`../codexctl`), которая в части бизнес-идеи остается действующей, за исключением подходов к реализации (там все планировалось делать через консольную утилиту, воркфлоу и лейблы, а тут полноценный сервис управления агентами, задачами и т.д.).

## Матрица чтения проектной документации (обязательна)

Перед началом работ по типу задачи читать минимум указанный набор:

| Тип задачи | Обязательные документы |
|---|---|
| Продуктовые требования/лейблы/этапы | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md` |
| Архитектура и модель данных | `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`, `docs/architecture/api_contract.md`, `docs/architecture/data_model.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/prompt_templates_policy.md` |
| Delivery/sprint/epics | `docs/delivery/development_process_requirements.md`, `docs/delivery/delivery_plan.md`, `docs/delivery/sprint_s*.md`, `docs/delivery/epic_s*.md`, `docs/delivery/epics/*.md` |
| Трассируемость | `docs/delivery/requirements_traceability.md`, `docs/delivery/issue_map.md`, `docs/delivery/sprint_s*.md`, `docs/delivery/epic_s*.md` |
| Ops и staging проверки | `.local/agents-temp-dev-rules.md`, `docs/ops/staging_runbook.md` |

## Архитектурные границы (обязательны)

- `services/external/*` = thin-edge:
  - HTTP ingress (webhooks/public endpoints), валидация, authn/authz, rate limiting, аудит, маршрутизация;
  - без доменной логики (use-cases) и без прямых postgres-репозиториев.
- `services/internal/*` = доменная логика и владельцы БД:
  - доменные модели/use-cases, репозитории, интеграции через интерфейсы/адаптеры;
  - внутреннее service-to-service взаимодействие через gRPC по контрактам в `proto/`.
- `services/jobs/*` = фоновые процессы и reconciliation (идемпотентно, состояние в БД).

## Транспортные контракты и модели (обязательны)

- При изменениях transport-слоя, DTO/кастеров и доменных моделей обязательно читать:
  - `docs/design-guidelines/go/services_design_requirements.md` (backend);
  - `docs/design-guidelines/vue/frontend_architecture.md` (frontend).
- В `transport/http|grpc` запрещены `map[string]any`/`[]any`/`any` как контракт ответа.
- Handlers возвращают только typed DTO-модели; маппинг transport <-> domain/proto выполняется через явные кастеры.
- Для `services/external/*` и `services/staff/*` действует contract-first OpenAPI:
  - любое изменение HTTP endpoint/DTO сначала в `api/server/api.yaml`;
  - затем регенерация backend/frontend codegen-артефактов;
  - merge запрещён, если маршруты/DTO в коде расходятся со спецификацией.
  - при любом изменении codegen-охвата (новый сервис/app или изменение путей/целей генерации) обязательно синхронно обновлять:
    - `Makefile` (`gen-openapi-*`);
    - `tools/codegen/**`;
    - `.github/workflows/contracts_codegen_check.yml`;
    - `docs/design-guidelines/go/code_generation.md`.
- Для HTTP DTO размещать модели и кастеры в `internal/transport/http/{models,casters}` (или эквивалентно по протоколу в рамках сервиса).
- Доменные типы размещать в `internal/domain/types/{entity,value,enum,query,mixin}`; не объявлять доменные модели ad-hoc в больших service/handler файлах.
- Маппинг ошибок выполняется только на границе транспорта (HTTP error handler / gRPC interceptor); в handlers запрещены локальные “переводы” ошибок между слоями.
- `context.Background()` создаётся только в composition root (`internal/app/*`); в transport/domain/repository-слоях использовать только прокинутый контекст.

## Образы сервисов (обязательны)

- В монорепо у каждого Go-сервиса собственный Dockerfile в `services/<zone>/<service>/Dockerfile`.
- У каждого frontend-сервиса обязателен `services/<zone>/<service>/Dockerfile` с минимум двумя target:
  - `dev` (staging/dev runtime);
  - `prod` (runtime на веб-сервере, например `nginx`, со статическим бандлом).
- Для каждого frontend-сервиса обязателен отдельный манифест в `deploy/base/<service>/*.yaml.tpl`.
- Раздутый “общий” Dockerfile для нескольких сервисов не используется как основной путь сборки/deploy.
- Для staging/CI обязательны раздельные image vars и image repositories на каждый deployable-сервис:
  - шаблон: `CODEXK8S_<SERVICE>_IMAGE`;
  - шаблон: `CODEXK8S_<SERVICE>_INTERNAL_IMAGE_REPOSITORY`.

## Порядок выкладки staging (обязателен)

- Применяется последовательность:
  `stateful dependencies -> migrations -> internal domain services -> edge services -> frontend`.
- Ожидание готовности зависимостей выполняется через `initContainers` в манифестах сервисов, а не через retry-циклы старта в Go-коде.

## Миграции и schema governance (обязательны)

- Миграции БД хранятся *внутри держателя схемы*:
  `services/<zone>/<db-owner-service>/cmd/cli/migrations/*.sql` (goose) согласно `docs/design-guidelines/go/*`.
- Shared DB без владельца запрещён: если БД общая, должен быть один сервис-владелец схемы и миграций.

## Внешние зависимости (обязательны)

- Любая новая внешняя зависимость (Go/TS) должна быть добавлена в
  `docs/design-guidelines/common/external_dependencies_catalog.md` вместе с обоснованием.
- Самописные “велосипеды” для типовых задач (например, форматирование дат) не добавлять, если есть утверждённая библиотека.

## Что считать источником правды

- Архитектурный стандарт: `docs/design-guidelines/**`.
- Целевая структура репозитория: `services/external|staff|internal|jobs|dev` + `libs` + `deploy` + `bootstrap` + `proto`.
- Оркестрация инфраструктуры: Kubernetes API через Go SDK (`client-go`), без shell-first подхода как основы.
- Интеграция с репозиториями: через интерфейсы провайдеров (`RepositoryProvider`),
  с текущей реализацией GitHub и заделом под GitLab.
- Модель процессов: webhook-driven, без GitHub Actions workflow как основного механизма выполнения.
- Хранилище сервиса: PostgreSQL (`JSONB` + `pgvector`) как единая точка синхронизации между pod'ами.
- MCP служебные ручки: встроенные Go-реализации в `codex-k8s`; `github.com/codex-k8s/yaml-mcp-server` остаётся расширяемым пользовательским слоем.
- Апрувы/экзекьюторы MCP: использовать универсальные HTTP-контракты (Telegram/Slack/Mattermost/Jira и др. как адаптеры), без вендорной привязки в core.
- Операционная продуктовая модель агентов/лейблов/этапов:
  `docs/product/agents_operating_model.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`.
- Процесс разработки и doc-governance: `docs/delivery/development_process_requirements.md`.

## Неподвижные ограничения продукта

- Поддерживается только Kubernetes.
- Регистрация пользователей отключена: вход через GitHub OAuth с матчингом по email,
  разрешённым администратором.
- Пользовательские настройки, шаблоны инструкций, сессии агентов, журналы действий,
  состояние слотов и рантаймов — в БД.
- Поддерживается learning mode: для задач пользователя добавляются explain-инструкции
  (почему/зачем/компромиссы), а после PR могут публиковаться образовательные комментарии.
- Секреты платформы и настройки деплоя `codex-k8s` берутся из env.
- Имена env/secrets/CI variables для платформы используют префикс `CODEXK8S_`
  (исключения допускаются только для внешних контрактов, например `POSTGRES_*` внутри контейнера PostgreSQL).
- Токены доступа к repo хранятся в БД в зашифрованном виде.

## Обязательные шаги перед PR

- Пройти `docs/design-guidelines/common/check_list.md`.
- Если затронут Go-код, пройти `docs/design-guidelines/go/check_list.md`.
- Если затронут Vue-код, пройти `docs/design-guidelines/vue/check_list.md`.
- Обновить документацию, если меняется поведение API, webhook-процессы,
  модель данных, RBAC, формат `services.yaml` или MCP-контракты.
