# Go Design Guidelines

Документы для Go backend.

- `docs/design-guidelines/go/check_list.md` — чек-лист перед PR для Go изменений.
- `docs/design-guidelines/go/services_design_requirements.md` — структура сервиса, домен/кастеры, repo+SQL правила, OpenAPI/AsyncAPI.
- `docs/design-guidelines/go/infrastructure_integration_requirements.md` — Postgres/Redis/секреты/миграции (goose) и запреты.
- `docs/design-guidelines/go/observability_requirements.md` — логи/трейсы/метрики (OTel/Jaeger/Prometheus).
- `docs/design-guidelines/go/protobuf_grpc_contracts.md` — правила gRPC `.proto` как транспортного контракта.
- `docs/design-guidelines/go/rest.md` — REST стек (echo + OpenAPI validation + codegen + swagger UI).
- `docs/design-guidelines/go/grpc.md` — gRPC (границы, контракты, ссылки на codegen).
- `docs/design-guidelines/go/websockets.md` — WebSocket (контракт AsyncAPI, правила сервера).
- `docs/design-guidelines/go/code_generation.md` — обязательные правила и команды кодогенерации.
- `docs/design-guidelines/go/code_commenting_rules.md` — правила комментариев в Go.
- `docs/design-guidelines/go/error_handling.md` — обязательные правила обработки ошибок в Go.
- `docs/design-guidelines/go/libraries.md` — что выносить в `libs/go/*` и как.
- `docs/design-guidelines/common/external_dependencies_catalog.md` — согласованный список внешних библиотек и инструментов.

Специфика `codex-k8s`:
- Kubernetes интеграция только через `client-go` и адаптеры.
- Репозитории (GitHub/GitLab) только через provider-интерфейсы.
- Оркестрация процессов event/webhook-driven, без workflow-first зависимостей.
- Состояние процессов и синхронизация pod'ов — через PostgreSQL (`JSONB` + `pgvector`).
- Планирование и закрытие daily задач в спринте — по
  `docs/delivery/development_process_requirements.md`.
