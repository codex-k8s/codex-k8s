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
    ├── controlplane/                                   internal gRPC client для platform-owned callback/state path
    ├── service/                                        Telegram transport/rendering/voice STT без platform semantics
    └── transport/http/                                 HTTP handlers/casters и health/metrics
```

Границы ответственности:
- принимает `worker -> adapter` delivery envelope `telegram-interaction-v1`;
- вызывает Telegram Bot API, принимает raw webhook и нормализует text/voice replies;
- конвертирует voice replies через `ffmpeg` и OpenAI STT перед отправкой platform-owned callback;
- принимает raw Telegram webhook, проверяет `X-Telegram-Bot-Api-Secret-Token`, делает `answerCallbackQuery`;
- пересылает normalized callback envelope напрямую в `control-plane` по internal gRPC, не владея platform semantics или БД платформы.
