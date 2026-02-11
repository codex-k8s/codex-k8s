# Go чек-лист перед PR

Используется как self-check перед созданием PR. В PR достаточно написать: «чек-лист выполнен, релевантно N пунктов, все выполнены».

## Архитектура и структура
- Структура сервиса соответствует `docs/design-guidelines/go/services_design_requirements.md` (domain/transport/repository разделены; нет доменной логики в transport).
- Доменные модели разложены системно (`internal/domain/types/{entity,value,enum,query,mixin}`), а не объявлены ad-hoc внутри service/handler файлов.
- В transport-слое ответы типизированы (DTO модели + кастеры); нет `map[string]any`/`[]any`/`any` как API-контрактов.
- JSON payload, сохраняемые в БД/события (например `*_payload`), типизированы через struct + caster; нет `map[string]any` в коммитнутом production-коде.
- Повторяющиеся строковые доменные значения вынесены в typed-константы (без копипасты литералов по коду).
- Ошибки маппятся на границе транспорта (HTTP error handler / gRPC interceptor); в handlers нет ad-hoc маппинга межслойных ошибок.
- `context.Background()` создан только в composition root (`internal/app/*`); в transport/domain/repository используется прокинутый контекст.
- Функции/методы оформлены с компактными сигнатурами (предпочтительно в одну строку); при большом числе аргументов используется `Config/Params/Input` структура.
- Интеграция с Kubernetes идёт через интерфейс/адаптер; прямой shell-first сценарий не является основным путём.
- Интеграция с репозиториями идёт через provider-интерфейсы (без GitHub-specific логики в домене).
- HTTP (если есть): OpenAPI в `api/server/api.yaml`; validation/codegen выполнены.
- Async/webhook payload contracts (если есть): описаны в `api/server/asyncapi.yaml`.
- Если менялись контракты (OpenAPI/proto/AsyncAPI): выполнена регенерация через `make`, и изменения в `**/generated/**` закоммичены.

## Webhook и процессы
- Процессы запускаются webhook-событиями и внутренними событиями из БД; workflow-first зависимость не добавлена.
- Для long-running задач есть идемпотентность, ретраи и запись состояния/блокировок в БД.
- Изменения lifecycle pod/slot фиксируются в таблицах состояния и журнале событий.

## Postgres и SQL
- Миграции (goose): `services/<zone>/<db-owner-service>/cmd/cli/migrations/*.sql` (timestamp; `-- +goose Up/Down`); история не переписывается.
- Repo интерфейсы в `internal/domain/repository/<model>/repository.go`; реализации в `internal/repository/postgres/<model>/repository.go`.
- SQL только в `internal/repository/postgres/<model>/sql/*.sql` + `//go:embed`; SQL-строки в Go запрещены.
- SQL-запросы именованы комментариями `-- name: <model>__<operation> :one|:many|:exec`.
- Для динамических данных используются `JSONB`; для векторного поиска корректно применён `pgvector`.

## Безопасность
- Секреты платформы читаются из env; не хардкодятся и не логируются.
- Repo-токены и чувствительные данные сохраняются только в шифрованном виде.
- OAuth/аутентификация не обходятся через debug или “временные” backdoor-механизмы.

## Автопроверки (обязательно перед PR)
- Соблюдён `docs/design-guidelines/go/code_commenting_rules.md`.
- В каждом изменённом Go-модуле выполнен `go mod tidy`.
- Если добавлена/обновлена внешняя Go библиотека, обновлён
  `docs/design-guidelines/common/external_dependencies_catalog.md`.
- Прогнан `make lint-go` и исправлены нарушения.
- Прогнан `make dupl-go`; дубли устранены или выделены в отдельную задачу.
