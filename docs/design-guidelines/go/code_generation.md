# Кодогенерация (контракты -> код)

Цель: после изменения контрактов (OpenAPI/proto/AsyncAPI) артефакты регенерируются через `make`, коммитятся и не правятся руками.

## Общие правила

- Любой сгенерированный код живет только в `**/generated/**`.
- Сгенерированное руками не правим.
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
- `@hey-api/openapi-ts` + `@hey-api/client-axios`

Выход:
- `src/shared/api/generated/**`

Запуск:
```bash
make gen-openapi-ts APP=services/<zone>/<app> SPEC=services/<zone>/<service>/api/server/api.yaml
```
