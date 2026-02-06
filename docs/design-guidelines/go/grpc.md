# gRPC в Go

Цель: единый подход к внутренним синхронным вызовам, контрактам и ошибкам.

## Контракт
- gRPC контракт — `.proto` в `proto/` (источник правды транспорта).
- Правила совместимости/версий: `docs/design-guidelines/go/protobuf_grpc_contracts.md`.

## Клиенты (вызовы других сервисов)
- gRPC клиенты к другим сервисам размещаем в `internal/clients/*` (реализация доменных портов).
- Подключаем клиентов в `internal/app` (composition root); домен не импортирует gRPC generated пакеты.

## Серверная граница
- gRPC методы (handlers) тонкие: принимают request, маппят в домен, вызывают доменный use-case/service, возвращают `error`.
- Маппинг доменных ошибок -> `codes.*` и recovery — в interceptors (см. `docs/design-guidelines/go/error_handling.md`).

## Codegen и dev-гейтвей
Генерация кода и dev-инфраструктуры (grpc-gateway + OpenAPI) описаны в:
- `docs/design-guidelines/go/code_generation.md`
