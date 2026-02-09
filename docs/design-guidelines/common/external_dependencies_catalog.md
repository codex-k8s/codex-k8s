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
| `github.com/getkin/kin-openapi` | n/a (planned in runtime usage) | OpenAPI validation | валидация request/response по контракту |
| `github.com/prometheus/client_golang` | `v1.23.2` | Observability | `/metrics` и базовые метрики сервиса |
| `github.com/jackc/pgx/v5` | `v5.8.0` | PostgreSQL driver | доступ к PostgreSQL |
| `github.com/google/uuid` | `v1.6.0` | Utility | генерация идентификаторов |
| `github.com/caarlos0/env/v11` | `v11.3.1` | Config | типобезопасный env->struct парсинг конфигурации |
| `k8s.io/client-go` | `v0.35.0` | Kubernetes integration | запуск/проверка Job через Kubernetes SDK |
| `k8s.io/api` | `v0.35.0` | Kubernetes API types | типы `batch/v1`, `core/v1` для Job/Pod |
| `k8s.io/apimachinery` | `v0.35.0` | Kubernetes API machinery | ошибки API, meta types, утилиты client-go |

## Frontend (Vue/TS) — planned baseline

| Dependency | Status | Scope | Why |
|---|---|---|---|
| `vue` | planned | UI framework | staff web-console |
| `vue-router` | planned | Routing | маршрутизация staff UI |
| `pinia` | planned | State management | глобальное состояние UI |
| `axios` | planned | HTTP client | вызовы staff/private API |

## Infrastructure and CI tools — in use

| Tool | Scope | Why |
|---|---|---|
| `gh` CLI | bootstrap scripts | настройка GitHub secrets/vars/webhooks |
| `kubectl` | bootstrap/deploy scripts | применение Kubernetes manifests |
| `helm` | bootstrap scripts | установка ARC и инфраструктурных компонентов |
| `openssl` | bootstrap scripts | генерация секретов |
| `kaniko` | CI build pipeline | сборка образа внутри кластера |

## Процесс изменений каталога

- PR с новой зависимостью должен обновлять:
  - этот файл;
  - релевантный гайд (`go/libraries.md`, `vue/libraries.md` и т.п.);
  - технические артефакты (`go.mod`, `package.json`, workflow/bootstrap при необходимости).
- Без обновления каталога изменение считается неполным.
