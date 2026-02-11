# Кодогенерация (контракты -> код)

Цель: после изменения контрактов (OpenAPI/proto/AsyncAPI) артефакты регенерируются через `make`, коммитятся и не правятся руками.

## Общие правила

- Любой сгенерированный код живет только в `**/generated/**`.
- Сгенерированное руками не правим.
- Изменение перечня сервисов/приложений, участвующих в codegen, должно синхронно обновлять:
  - цели `gen-openapi-*` в `Makefile`;
  - конфиги codegen в `tools/codegen/**`;
  - CI-проверку `.github/workflows/contracts_codegen_check.yml`.
- Источник правды транспорта:
  - REST: `api/server/api.yaml` (OpenAPI YAML)
  - gRPC: `proto/**/*.proto`
  - async/webhook: `api/server/asyncapi.yaml` (если используется)

## OpenAPI (REST) -> Go

Инструмент:
- `github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`

Выход:
- `internal/transport/http/generated/openapi.gen.go`

Запуск:
```bash
make gen-openapi-go SVC=services/<zone>/<service>
```

Обязательный make-блок:
- `gen-openapi-go` — генерация Go transport-артефактов;
- `gen-openapi-ts` — генерация TS-клиента для frontend;
- `gen-openapi` — агрегирующая цель для CI и локальной проверки.

## Protobuf/gRPC -> Go

Инструменты:
- `google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- `google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`

Выход:
- `internal/transport/grpc/generated/**`

Запуск:
```bash
make gen-proto-go SVC=services/<zone>/<service>
```

## AsyncAPI (webhook/event payloads)

Контракт:
- `api/server/asyncapi.yaml`

Применение в `codex-k8s`:
- описание webhook payloads и внутренних async-событий,
- опциональная генерация моделей для transport-слоя.

Валидация:
```bash
make validate-asyncapi SVC=services/<zone>/<service>
```

## Frontend codegen по OpenAPI (TypeScript + Axios)

Рекомендуемый инструмент:
- `@hey-api/openapi-ts` (клиенты `@hey-api/client-*` встроены начиная с `v0.73.0`; отдельная установка `@hey-api/client-axios` не требуется)

Выход:
- `src/shared/api/generated/**`

Запуск:
```bash
make gen-openapi-ts APP=services/<zone>/<app> SPEC=services/<zone>/<service>/api/server/api.yaml
```

## Проверка консистентности generated-кода в CI

- Обязательная проверка: `.github/workflows/contracts_codegen_check.yml`.
- Workflow должен выполнять:
  - установку зависимостей frontend;
  - `make gen-openapi`;
  - `git diff --exit-code` по OpenAPI-generated артефактам.
- Любое расширение/изменение codegen-охвата (новый backend service или frontend app) сопровождается правкой этого workflow.
