# control-plane

`control-plane` — внутренний доменный сервис платформы: orchestrates use-cases, владеет схемой БД и политиками выполнения.

```text
services/internal/control-plane/                     доменный backend (владелец схемы и use-case логики)
├── README.md                                        карта структуры сервиса и ключевых областей
├── Dockerfile                                       сборка runtime-образа сервиса
├── cmd/
│   ├── control-plane/main.go                        composition root запуска gRPC/MCP/внутренних контуров
│   └── cli/migrations/                              миграции БД (schema governance этого сервиса)
└── internal/
    ├── app/                                         конфигурация, bootstrap и жизненный цикл приложения
    ├── clients/                                     инфраструктурные адаптеры внешних API (GitHub/Kubernetes)
    ├── domain/                                      доменные use-cases и типы; ОБЯЗАТЕЛЬНО смотреть при изменении бизнес-логики
    │   ├── agentcallback/                           обработка callback из agent runtime
    │   ├── mcp/                                     доменная оркестрация MCP tools/policy
    │   ├── repository/                              контракты provider/repository для доменного слоя
    │   │   └── runtimedeploytask/                   persisted desired/actual state контур full-env deploy
    │   ├── runstatus/                               use-cases статусов run и state transitions
    │   ├── runtimedeploy/                           декларативный full-env deploy/reconcile из `services.yaml`
    │   ├── staff/                                   внутренние staff use-cases управления платформой
    │   ├── types/                                   доменные entity/value/enum/query типы
    │   └── webhook/                                 обработка webhook-driven сценариев
    ├── repository/postgres/                         PostgreSQL-реализации доменных репозиториев
    │   └── runtimedeploytask/                       lease-aware очередь runtime deploy задач в БД
    └── transport/                                   транспортные адаптеры сервиса
        ├── grpc/                                    внутренний gRPC API
        ├── mcp/                                     MCP StreamableHTTP/control tools endpoint
        └── agentcallback/                           callback transport для agent runner
```
