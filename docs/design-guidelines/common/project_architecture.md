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
  - Рендер и применение выполняются Go-компонентами (`cmd/codex-bootstrap`, `runtime-deploy`); shell-скрипты
    не являются основным механизмом deploy/sync.
  - Для production порядок выкладки задаётся по слоям:
    `stateful dependencies -> migrations -> internal domain services -> edge services -> frontend`.
    Ожидание зависимостей оформляется через `initContainers` в манифестах сервисов.
  - Для monorepo multi-service deploy используются раздельные образы/репозитории для каждого
    deployable-сервиса (шаблон нейминга: `CODEXK8S_<SERVICE>_IMAGE`,
    `CODEXK8S_<SERVICE>_INTERNAL_IMAGE_REPOSITORY`).
- `bootstrap/` — скрипты bootstrap (готовый кластер или установка k3s).
- `docs/` — документация и решения.
- `tools/` — утилиты и генерация.

### Изменение состава deployable-сервисов (обязательно синхронно)

Если в монорепо добавляется/удаляется deployable-сервис, либо меняются его docker context / Dockerfile / image repository:
- обновлять сборку/зеркалирование образов в runtime deploy:
  `services/internal/control-plane/internal/domain/runtimedeploy/service_build.go`,
  `services/internal/control-plane/internal/domain/runtimedeploy/service_reconcile.go`,
  `deploy/base/kaniko/kaniko-build-job.yaml.tpl`,
  `deploy/base/kaniko/mirror-image-job.yaml.tpl`;
- обновлять production deploy env/vars и манифесты:
  `services/internal/control-plane/cmd/runtime-deploy/main.go`,
  `deploy/base/codex-k8s/codegen-check-job.yaml.tpl`,
  `services.yaml`;
- обновлять bootstrap-синхронизацию GitHub vars/secrets и секретов Kubernetes:
  `bootstrap/host/config.env.example`,
  `bootstrap/host/bootstrap_remote_production.sh`,
  `cmd/codex-bootstrap/internal/cli/github_sync.go`,
  `cmd/codex-bootstrap/internal/cli/sync_secrets.go`,
  `cmd/codex-bootstrap/internal/cli/cleanup.go`;
- обновлять deploy manifests сервиса в `deploy/base/<service>/*.yaml.tpl`
  и Dockerfile сервиса в `services/<zone>/<service>/Dockerfile`.
- если меняется набор инструментов для agent-runner (dogfooding), обновлять
  `services/jobs/agent-runner/scripts/bootstrap_tools.sh` и документировать изменения в релевантном эпике/операционном runbook.

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
- Для каждого frontend-сервиса обязателен собственный Dockerfile с target `dev` и `prod`,
  а также отдельный deploy manifest в `deploy/base/<service>/*.yaml.tpl`.

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
