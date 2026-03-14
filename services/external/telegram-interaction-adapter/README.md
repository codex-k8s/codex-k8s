# telegram-interaction-adapter

`telegram-interaction-adapter` — внешний edge-сервис платформы для Telegram-specific delivery/webhook path поверх typed interaction contract Sprint S11.

```text
services/external/telegram-interaction-adapter/         deployable Telegram adapter contour
├── README.md                                           карта структуры сервиса и runtime-boundary
├── Dockerfile                                          сборка runtime-образа сервиса
├── api/
│   └── server/
│       └── api.yaml                                    OpenAPI source of truth для delivery/webhook HTTP-контрактов
├── cmd/
│   └── telegram-interaction-adapter/
│       └── main.go                                     composition root запуска сервиса
└── internal/
    ├── app/                                            конфиг и bootstrap
    ├── service/                                        Telegram transport/rendering/session state без platform semantics
    └── transport/http/                                 HTTP handlers/casters и health/metrics
```

Границы ответственности:
- принимает `worker -> adapter` delivery envelope `telegram-interaction-v1`;
- вызывает Telegram Bot API, сохраняет локальное encrypted session state для callback handles/tokens;
- принимает raw Telegram webhook, проверяет `X-Telegram-Bot-Api-Secret-Token`, делает `answerCallbackQuery`;
- пересылает только normalized callback envelope в `api-gateway`, не владея platform semantics или БД платформы.
