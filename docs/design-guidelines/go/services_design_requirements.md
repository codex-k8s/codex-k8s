# Сервисы: требования к проектированию

Цель: все сервисы `codex-k8s` устроены единообразно, с явными доменными границами, интерфейсной интеграцией и наблюдаемостью по умолчанию.

## Сервис: ответственность, имя и размещение

Имена:
- `kebab-case`.
- Имя отражает домен/роль, а не технологию.

Размещение:
- `services/<zone>/<service-name>/`, где `<zone>` ∈ `internal|external|staff|jobs|dev`.

## Рекомендуемое ядро codex-k8s

- `services/internal/control-plane` — домен платформы (проекты, репозитории, агенты, слоты, webhook orchestration, аудит).
- `services/external/api-gateway` — webhook/API входы.
- `services/staff/web-console` — frontend (Vue3).
- `services/jobs/worker` — фоновые задачи и reconciliation.
- `services/dev/webhook-simulator` — dev-only инструменты.

## Выбор протокола

### gRPC (внутренний sync)
- Контракты в `proto/`, совместимость обязательна.
- Используется для service-to-service взаимодействия внутри платформы.

### HTTP/REST (внешний и staff API)
- Публичные/внутренние API для webhook и UI.
- OpenAPI YAML обязателен для `external|staff`.

### WebSocket (опционально)
- Только где есть реальный realtime use-case в staff UI.
- Контракт сообщений фиксируется в AsyncAPI.

## Внутренняя структура Go-сервиса

Внутри `services/<zone>/<service-name>/`:

- `cmd/<service-name>/main.go` — thin entrypoint.
- `internal/app/` — composition root + lifecycle + graceful shutdown.
- `internal/transport/{http,grpc,ws}/` — handlers и middleware, без доменной логики.
- `internal/transport/{http,grpc,ws}/models/` — typed DTO контракты конкретного транспорта.
- `internal/transport/{http,grpc,ws}/casters/` — явный маппинг transport DTO <-> domain/proto.
- `internal/domain/` — бизнес-правила, модели, use-cases, порты.
  - `internal/domain/service/` — доменная бизнес-логика (use-cases).
  - `internal/domain/errs/` — доменные typed errors (если нужны).
  - `internal/domain/casters/` — маппинг persistence <-> domain (без transport/pgx зависимостей).
  - `internal/domain/helpers/` — локальные доменные helpers (валидация, нормализация, конвертеры).
  - `internal/domain/types/` — доменные типы, разнесённые по категориям:
    - `internal/domain/types/entity/*.go` — сущности;
    - `internal/domain/types/value/*.go` — value objects;
    - `internal/domain/types/enum/*.go` — enum-подобные типы;
    - `internal/domain/types/query/*.go` — фильтры/параметры use-case;
    - `internal/domain/types/mixin/*.go` — общие встраиваемые фрагменты (paging/time-range).
- `internal/domain/repository/<model>/repository.go` — интерфейсы репозиториев.
- `internal/repository/postgres/<model>/repository.go` — реализации репозиториев.
- `internal/repository/postgres/<model>/sql/*.sql` — SQL (через `//go:embed`).
  - запросы именуются комментариями в стиле
    `-- name: <model>__<operation> :one|:many|:exec`
    для стабильной привязки SQL к коду;
- `internal/clients/kubernetes/` — адаптеры Kubernetes SDK.
- `internal/clients/repository/` — адаптеры provider-интерфейсов (`github`, позже `gitlab`).
- `internal/observability/` — подключение логов/трейсов/метрик.
- `cmd/cli/migrations/*.sql` — миграции БД (goose) для этого сервиса, если он держатель схемы.
  - В монорепо это означает путь:
    `services/<zone>/<service-name>/cmd/cli/migrations/*.sql`.
- `api/server/api.yaml` — OpenAPI.
- `api/server/asyncapi.yaml` — async/webhook/event контракты (если используются).
- `internal/transport/*/generated/**` — только сгенерированный код.

## Требования к transport DTO и кастерам

- Контракт ответа transport-ручек должен быть строго типизирован.
- Для HTTP/gRPC handlers запрещено возвращать `map[string]any`, `[]any`, `any`.
- Маппинг между transport DTO, proto и доменными моделями должен быть вынесен в отдельные `casters`.
- Handler’ы не содержат доменных преобразований “вручную” и не агрегируют произвольные структуры ad-hoc.
- Для JSON payload в БД/событиях/очередях (например `run_payload`, `flow_events.payload`) использовать typed structs и кастеры; `map[string]any` допускается только как временная локальная отладка и не коммитится.

## Сигнатуры функций и параметры

- Сигнатуры функций и методов писать в одну строку, если это технически возможно без потери читаемости.
- Если аргументов много, вводить входную `Params/Config/Input` структуру вместо длинного списка аргументов.
- Для конструкторов сервисов использовать конфиг-структуры (`Config`) вместо позиционных аргументов.

## Строковые доменные значения

- Повторяющиеся доменные строковые значения выносить в константы.
- Для закрытых наборов строк использовать typed aliases (`type ... string`) и константы поверх них.
- В production-коде запрещено размножать одинаковые литералы статусов/типов событий в разных пакетах.

## Доменные контексты (минимум)

В `internal/control-plane/internal/domain/` должны быть отдельные bounded contexts:
- `users` (OAuth-сессии, доступы)
- `projects` (проекты и membership)
- `repositories` (repo bindings и токены)
- `webhooks` (ingest/validation/event mapping)
- `agents` (шаблоны инструкций, профили агентов)
- `agent_runs` (сессии, статусы, токены)
- `slots` (распределение/блокировки)
- `docs_kb` (шаблоны, метаданные, чанки, индексация)
- `audit` (журнал действий и событий)

## Нефункциональные требования

Обязательное:
- Health: `/health/livez`, `/health/readyz`.
- Metrics: `/metrics`.
- Structured logs, без секретов/PII.
- OTel tracing и пропагация контекста.
- Graceful shutdown по `SIGINT|SIGTERM|SIGQUIT|SIGHUP`.
- Базовый контекст приложения (`context.Background()`) создаётся только в `internal/app/*` (composition root) и прокидывается зависимостям через конструкторы/методы; в transport/domain/repository-слоях прямой вызов `context.Background()` запрещён.

## Запрещено

- Доменная логика в `transport/*`.
- Прямой импорт `client-go`/`go-github` в домене.
- SQL строками в Go-коде.
- Обход interface-layer и вызов vendor SDK из use-case слоя.
- Использование `map[string]any`/`any` как публичного transport-контракта.
