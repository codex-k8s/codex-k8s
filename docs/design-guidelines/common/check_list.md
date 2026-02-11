# Общий чек-лист перед PR

Используется как self-check перед созданием PR. В PR достаточно написать: «чек-лист выполнен, релевантно N пунктов, все выполнены».

## Общее
- Не размыты доменные границы: один сервис/приложение = один bounded context; нет “service-олигарха”.
- Зона выбрана корректно: `internal|external|staff|jobs|dev` (см. `docs/design-guidelines/common/project_architecture.md`).
- Для `external|staff` edge остаётся thin-edge (валидация/auth/маршрутизация), без доменной логики backend.
- Контракты транспорта не “вшиты в код строками” и имеют источник правды:
  - gRPC: `proto/` (версионирование/совместимость)
  - HTTP: OpenAPI YAML
  - async/webhook payloads: AsyncAPI YAML (если используются)
- Секреты не хардкодятся и не коммитятся; в логах нет секретов/PII.
- Имена platform env/secrets/CI variables унифицированы с префиксом `CODEXK8S_`;
  имена без префикса `CODEXK8S_` не добавляются.
- Kubernetes манифесты не “вшиты” heredoc’ами в bash: шаблоны лежат в `deploy/base/**`,
  а `deploy/scripts/**` только рендерит и применяет их.
- Для multi-service deploy у каждого deployable-сервиса есть собственные image vars/repositories
  (шаблон нейминга: `CODEXK8S_<SERVICE>_IMAGE` и `CODEXK8S_<SERVICE>_INTERNAL_IMAGE_REPOSITORY`).
- Порядок выкладки staging задан явно и соблюдён:
  stateful dependencies -> migrations -> internal domain services -> edge services -> frontend;
  ожидание зависимостей выполнено через `initContainers` в Kubernetes-манифестах.
- Вынос общего кода в `libs/*` оправдан (>= 2 потребителя); нет “god-lib”.
- Если добавлена/обновлена внешняя зависимость, обновлён
  `docs/design-guidelines/common/external_dependencies_catalog.md`.

## Специфика codex-k8s
- Поддержка оркестрации ограничена Kubernetes; нет кода под другие оркестраторы.
- Kubernetes-операции выполняются через SDK/интерфейсы (не shell-first как основной путь).
- Интеграция с репозиториями идет через интерфейс провайдера; GitHub-специфика не просачивается в домен.
- Процессы запускаются webhook-событиями; workflow-first сценарии не добавлены в обход архитектуры.
- Состояние long-running процессов, слотов, агентных запусков и блокировок хранится в PostgreSQL.
- Данные, требующие гибкой структуры, хранятся в `JSONB`; векторный поиск использует `pgvector`.
- Секреты платформы читаются из env; repo-токены хранятся шифрованно.

## Профильные чек-листы
- Если PR затрагивает Go: выполнен `docs/design-guidelines/go/check_list.md`.
- Если PR затрагивает Vue: выполнен `docs/design-guidelines/vue/check_list.md`.
- Если PR затрагивает Go: выполнен `go mod tidy` в изменённых модулях и прогнаны `make lint-go` и `make dupl-go` (или `make lint`) с устранением нарушений.
