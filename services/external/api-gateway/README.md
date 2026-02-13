# api-gateway

`api-gateway` — edge-сервис: принимает внешние webhook/API-запросы, валидирует и маршрутизирует их во внутренние доменные сервисы.

```text
services/external/api-gateway/                       thin-edge сервис без доменной логики
├── README.md                                        карта структуры сервиса и обязательных точек входа
├── Dockerfile                                       сборка runtime-образа сервиса
├── api/                                             contract-first спецификации транспорта; ОБЯЗАТЕЛЬНО смотреть при изменениях HTTP/async контрактов
│   └── server/
│       ├── api.yaml                                 OpenAPI source of truth для external/staff HTTP endpoint'ов
│       └── asyncapi.yaml                            AsyncAPI описание webhook/event payload
├── cmd/
│   └── api-gateway/
│       └── main.go                                  composition root запуска сервиса
└── internal/
    ├── app/                                         wiring приложения, конфиг и bootstrap компонентов
    ├── auth/                                        authn/authz и работа с identity на edge-границе
    ├── controlplane/                                адаптер клиента во внутренний control-plane
    └── transport/http/                              HTTP handlers/DTO/casters и middleware транспорта
```
