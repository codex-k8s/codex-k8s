# Инструкции для ИИ-агентов (обязательно)

## Назначение

Этот файл задает обязательные правила работы с репозиторием `codex-k8s`.
Цель: вести код и документацию к целевой архитектуре cloud-сервиса,
который оркестрирует агентные процессы в Kubernetes и заменяет связку
`codexctl` + часть сценариев `yaml-mcp-server`.

## Главные правила

- Перед изменениями читать `docs/design-guidelines/AGENTS.md`.
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
- Коммиты/ветки и PR не делать, если об этом не было явных указаний в запросе пользователя.

## Архитектурные границы (обязательны)

- `services/external/api-gateway` = thin-edge:
  - HTTP ingress (webhooks/public endpoints), валидация, authn/authz, rate limiting, аудит, маршрутизация;
  - без доменной логики (use-cases) и без прямых postgres-репозиториев.
- `services/internal/control-plane` = доменная логика и владелец БД:
  - доменные модели/use-cases, репозитории, интеграции через интерфейсы/адаптеры;
  - внутреннее service-to-service взаимодействие через gRPC по контрактам в `proto/`.
- `services/jobs/worker` = фоновые процессы и reconciliation (идемпотентно, состояние в БД).

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
