# worker

`worker` — фоновый сервис очередей и reconciliation: исполняет отложенные задачи, синхронизирует состояние и обслуживает lifecycle run.

```text
services/jobs/worker/                                фоновые jobs и reconciliation контур
├── README.md                                        карта структуры worker-сервиса
├── Dockerfile                                       image фонового исполнителя
├── cmd/worker/main.go                               точка входа worker-процесса
└── internal/
    ├── app/                                         конфигурация и bootstrap worker runtime
    ├── clients/kubernetes/                          клиентские адаптеры Kubernetes API
    ├── controlplane/client.go                       клиент внутренних control-plane API
    ├── domain/                                      доменные контракты и use-cases worker-задач
    │   ├── repository/                              интерфейсы репозиториев для фоновых сценариев
    │   ├── types/                                   доменные типы worker-контекста
    │   └── worker/                                  бизнес-логика очередей, retries и cleanup
    └── repository/postgres/                         PostgreSQL-адаптеры хранения состояния jobs
```
