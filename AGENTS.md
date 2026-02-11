# Инструкции для ИИ-агентов (обязательно)

## Назначение

Этот файл задает обязательные правила работы с репозиторием `codex-k8s`.
Цель: вести код и документацию к целевой архитектуре cloud-сервиса,
который оркестрирует агентные процессы в Kubernetes и заменяет связку
`codexctl` + часть сценариев `yaml-mcp-server`.

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
  - сверить задачу с `docs/product/requirements_machine_driven.md`;
  - только после этого планировать и править код.
- Не редактировать сами гайды без явной задачи на изменение стандартов.

## Архитектурные границы (обязательны)

- `services/external/api-gateway` = thin-edge:
  - HTTP ingress (webhooks/public endpoints), валидация, authn/authz, rate limiting, аудит, маршрутизация;
  - без доменной логики (use-cases) и без прямых postgres-репозиториев.
- `services/internal/control-plane` = доменная логика и владелец БД:
  - доменные модели/use-cases, репозитории, интеграции через интерфейсы/адаптеры;
  - внутреннее service-to-service взаимодействие через gRPC по контрактам в `proto/`.
- `services/jobs/worker` = фоновые процессы и reconciliation (идемпотентно, состояние в БД).

## Транспортные контракты и модели (обязательны)

- В `transport/http|grpc` запрещены `map[string]any`/`[]any`/`any` как контракт ответа.
- Handlers возвращают только typed DTO-модели; маппинг transport <-> domain/proto выполняется через явные кастеры.
- Для HTTP DTO размещать модели и кастеры в `internal/transport/http/{models,casters}` (или эквивалентно по протоколу в рамках сервиса).
- Маппинг ошибок выполняется только на границе транспорта (HTTP error handler / gRPC interceptor); в handlers запрещены локальные “переводы” ошибок между слоями.

## Образы сервисов (обязательны)

- В монорепо у каждого Go-сервиса собственный Dockerfile в `services/<zone>/<service>/Dockerfile`.
- Раздутый “общий” Dockerfile для нескольких сервисов не используется как основной путь сборки/deploy.
- Для staging/CI обязательны раздельные image vars и image repositories:
  - `CODEXK8S_API_GATEWAY_IMAGE`, `CODEXK8S_CONTROL_PLANE_IMAGE`, `CODEXK8S_WORKER_IMAGE`
  - `CODEXK8S_API_GATEWAY_INTERNAL_IMAGE_REPOSITORY`, `CODEXK8S_CONTROL_PLANE_INTERNAL_IMAGE_REPOSITORY`, `CODEXK8S_WORKER_INTERNAL_IMAGE_REPOSITORY`
- Legacy fallback `CODEXK8S_IMAGE` допускается только как временная совместимость; запрещено оставлять multi-service конфигурацию в состоянии, где все сервисы фактически публикуются в один legacy-репозиторий.

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
- MCP служебные ручки: встроенные Go-реализации в `codex-k8s`; `yaml-mcp-server` остаётся расширяемым пользовательским слоем.
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
