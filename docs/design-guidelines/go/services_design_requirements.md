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
- `internal/domain/` — бизнес-правила, модели, use-cases, порты.
- `internal/domain/repository/<model>/repository.go` — интерфейсы репозиториев.
- `internal/repository/postgres/<model>/repository.go` — реализации репозиториев.
- `internal/repository/postgres/<model>/sql/*.sql` — SQL (через `//go:embed`).
- `internal/clients/kubernetes/` — адаптеры Kubernetes SDK.
- `internal/clients/repository/` — адаптеры provider-интерфейсов (`github`, позже `gitlab`).
- `internal/observability/` — подключение логов/трейсов/метрик.
- `cmd/cli/migrations/*.sql` — миграции БД (goose).
- `api/server/api.yaml` — OpenAPI.
- `api/server/asyncapi.yaml` — async/webhook/event контракты (если используются).
- `internal/transport/*/generated/**` — только сгенерированный код.

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

## Запрещено

- Доменная логика в `transport/*`.
- Прямой импорт `client-go`/`go-github` в домене.
- SQL строками в Go-коде.
- Обход interface-layer и вызов vendor SDK из use-case слоя.
