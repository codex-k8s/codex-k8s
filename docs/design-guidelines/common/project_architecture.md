# Архитектура проекта

Цель: предсказуемое развитие `codex-k8s` как централизованного control-plane сервиса для агентных процессов в Kubernetes.

База: DDD (bounded contexts) + Clean Architecture (зависимости “снаружи внутрь”) + единый инвентарь деплоя (`services.yaml`) + единый каркас директорий.

## Архитектурные ограничения codex-k8s

- Оркестратор: только Kubernetes.
- Интеграция с Kubernetes: Go SDK (`client-go`) через интерфейсы и адаптеры.
- Интеграция с репозиториями: интерфейсы провайдеров (`github` сейчас, `gitlab` позже).
- Процессы: webhook-driven (GitHub webhooks/внутренние события), без workflow-first модели.
- Хранилище и синхронизация multi-pod: PostgreSQL (`JSONB` + `pgvector`).
- MCP служебные ручки: реализуются в Go внутри `codex-k8s`.

## Структура репозитория

Верхний уровень:
- `services.yaml` — инвентарь деплоя и окружений.
- `services/` — сервисы по зонам (`internal|external|staff|jobs|dev`).
- `libs/` — переиспользуемый код (`go|ts|vue`).
- `proto/` — gRPC контракты (single source of truth для внутреннего sync API).
- `deploy/` — Kubernetes манифесты и overlays.
  - Манифесты и шаблоны YAML (`*.yaml.tpl`) живут в `deploy/base/**`.
  - Bash-скрипты в `deploy/scripts/**` не должны содержать “встроенные” multi-line YAML/JSON манифесты через heredoc.
    Скрипты только рендерят и применяют файлы из `deploy/base/**`.
  - Для monorepo multi-service deploy используются раздельные образы/репозитории по сервисам
    (`api-gateway`, `control-plane`, `worker`, `web-console`), а не единый legacy image.
- `bootstrap/` — скрипты bootstrap (готовый кластер или установка k3s).
- `docs/` — документация и решения.
- `tools/` — утилиты и генерация.

Рекомендуемое ядро сервиса:
- `services/internal/control-plane` — доменная логика платформы (проекты, репозитории, агенты, слоты, webhook orchestration, audit).
- `services/external/api-gateway` — внешний API для webhook/публичных интеграций.
- `services/staff/web-console` — UI (Vue3) для админов/пользователей платформы.
- `services/jobs/worker` — фоновые jobs/reconciliation/ротация токенов/индексация.
- `services/dev/webhook-simulator` — dev-only утилиты.

## Зоны сервисов: internal / external / staff / jobs / dev

### `services/internal/`
- Доменные правила платформы.
- Работа с БД, Kubernetes и repository providers через интерфейсы.
- Нет публичного ingress для бизнес-эндпоинтов.

### `services/external/`
- Публичные webhook/API точки входа.
- Валидация подписи, authn/authz, rate limiting, аудит.
- Без доменной логики orchestration внутри transport слоя.

### `services/staff/`
- Внутренняя консоль (Vue3).
- Доступ через GitHub OAuth и внутреннюю матрицу прав проекта.

### `services/jobs/`
- Async/фоновые процессы: reconciliation, ретраи, cleanups, ротации, индексация.
- Идемпотентность и устойчивость обязательны.

### `services/dev/`
- Только dev-инструменты.
- Не деплоятся в production.

## Границы ответственности

Правила:
- Один сервис = один bounded context и одна причина для изменения.
- Домен не зависит от transport/DB SDK напрямую.
- Kubernetes/GitHub/GitLab детали изолированы в адаптерах.
- Shared DB без владельца контекста запрещён; таблицы и данные имеют явного владельца.

## Схема взаимодействия (высокоуровнево)

1. Внешний webhook приходит в `external/api-gateway`.
2. Событие проходит валидацию и передаётся в `internal/control-plane`.
3. `control-plane` пишет состояние/события в PostgreSQL и ставит задачи в `jobs/worker`.
4. `jobs/worker` выполняет действия в Kubernetes/repo providers через интерфейсы и фиксирует результат в БД.
5. `staff/web-console` читает состояние и управляет системой через API.
