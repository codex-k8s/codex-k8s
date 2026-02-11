# External Dependencies Catalog

Назначение: единая точка, где фиксируются внешние библиотеки и инструменты,
разрешённые/используемые в `codex-k8s`.

## Правила ведения

- Любая новая внешняя зависимость сначала добавляется в этот каталог.
- Для каждой зависимости фиксируются:
  - где используется;
  - зачем нужна;
  - есть ли альтернатива;
  - кто владелец решения (роль/команда).
- Для Go зависимости версия фиксируется в `go.mod`; для JS/Vue — в `package.json`.
- Если зависимость удалена, запись не удаляется молча, а переводится в `deprecated` с датой.

## Backend (Go) — in use

| Dependency | Version | Scope | Why |
|---|---|---|---|
| `github.com/labstack/echo/v5` | `v5.0.3` | HTTP transport | единый REST стек для gateway/staff API |
| `github.com/getkin/kin-openapi` | `v0.133.0` | OpenAPI validation | загрузка/валидация OpenAPI и runtime request-validation в `api-gateway` |
| `github.com/oapi-codegen/runtime` | `v1.1.2` | OpenAPI generated transport runtime | типы/утилиты для сгенерированного OpenAPI Go-кода |
| `github.com/prometheus/client_golang` | `v1.23.2` | Observability | `/metrics` и базовые метрики сервиса |
| `github.com/jackc/pgx/v5` | `v5.8.0` | PostgreSQL driver | доступ к PostgreSQL |
| `github.com/google/uuid` | `v1.6.0` | Utility | генерация идентификаторов |
| `github.com/caarlos0/env/v11` | `v11.3.1` | Config | типобезопасный env->struct парсинг конфигурации |
| `github.com/golang-jwt/jwt/v5` | `v5.3.0` | Auth | выпуск и валидация short-lived JWT для staff API |
| `k8s.io/client-go` | `v0.35.0` | Kubernetes integration | запуск/проверка Job через Kubernetes SDK |
| `k8s.io/api` | `v0.35.0` | Kubernetes API types | типы `batch/v1`, `core/v1` для Job/Pod |
| `k8s.io/apimachinery` | `v0.35.0` | Kubernetes API machinery | ошибки API, meta types, утилиты client-go |
| `github.com/google/go-github/v82` | `v82.0.0` | Repository provider (GitHub) | настройка вебхуков и валидация доступа к репозиториям через GitHub API v3 |
| `github.com/google/go-querystring` | `v1.2.0` | Dependency of go-github | сериализация query params для GitHub API клиента |
| `google.golang.org/grpc` | `v1.78.0` | Internal transport | внутреннее service-to-service взаимодействие (`api-gateway` -> `control-plane`) |
| `google.golang.org/protobuf` | `v1.36.10` | Internal contracts | protobuf runtime для gRPC контрактов и сгенерированного кода в `proto/gen/go/**` |

## Frontend (Vue/TS) — in use

| Dependency | Status | Scope | Why |
|---|---|---|---|
| `vue` | in use (package.json) | UI framework | staff web-console |
| `vue-router` | in use (package.json) | Routing | маршрутизация staff UI |
| `pinia` | in use (package.json) | State management | минимальное состояние UI |
| `axios` | in use (package.json) | HTTP client | вызовы staff/private API |
| `vue-i18n` | in use (package.json) | i18n | все пользовательские тексты через i18n ключи |
| `vue3-cookies` | in use (package.json) | Cookies | хранение UI-настроек (например, язык) и единый cookie-адаптер |
| `date-fns` | in use (package.json) | Datetime formatting | безопасное форматирование дат/времени без самописных helpers |
| `@hey-api/openapi-ts` | in use (devDependency, `v0.92.3`) | OpenAPI codegen (TS) | генерация typed API-клиента для frontend из `api.yaml` |
| `@hey-api/client-axios` | deprecated (bundled in `@hey-api/openapi-ts` since `v0.73.0`) | OpenAPI axios client plugin | отдельная установка не требуется, использовать встроенный плагин через конфиг `openapi-ts` |

## Infrastructure and CI tools — in use

| Tool | Scope | Why |
|---|---|---|
| `gh` CLI | bootstrap scripts | настройка GitHub secrets/vars/webhooks |
| `kubectl` | bootstrap/deploy scripts | применение Kubernetes manifests |
| `helm` | bootstrap scripts | установка ARC и инфраструктурных компонентов |
| `openssl` | bootstrap scripts | генерация секретов |
| `kaniko` | CI build pipeline | сборка образа внутри кластера |
| `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen` | Make codegen pipeline | генерация Go transport-артефактов из OpenAPI |

## Процесс изменений каталога

- PR с новой зависимостью должен обновлять:
  - этот файл;
  - релевантный гайд (`go/libraries.md`, `vue/libraries.md` и т.п.);
  - технические артефакты (`go.mod`, `package.json`, workflow/bootstrap при необходимости).
- Без обновления каталога изменение считается неполным.
