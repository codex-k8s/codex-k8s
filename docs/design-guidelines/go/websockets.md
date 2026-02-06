# WebSocket в Go

Цель: единый подход к realtime каналам, контрактам сообщений и обработке ошибок.

## Контракт (AsyncAPI)

- Контракт WebSocket сообщений описывается в AsyncAPI YAML: `api/server/asyncapi.yaml`.
- В AsyncAPI фиксируем каналы, типы сообщений, payload schemas, версии и correlation поля.

## Типы сообщений

- Если используем генерацию по AsyncAPI: модели генерируются в `internal/transport/async/generated/**`.
- Сгенерированные типы руками не правим; маппинг в домен через casters.

## Серверный слой

Правила:
- WS handlers тонкие: handshake/auth, парсинг, маппинг, вызов домена.
- Бизнес-правила в обработчиках сообщений запрещены.
- Heartbeat обязателен (ping/pong или app-level), таймауты и лимиты соединений обязательны.
- Origin-policy обязательна: default same-origin, расширение только allowlist через env.

## Мосты/форвардеры событий

- Фоновая доставка событий в WS реализуется в transport-слое.
- `internal/app` только запускает процесс и управляет lifecycle через `ctx`.

## Observability

- Логировать connect/disconnect, ошибки парсинга/валидации, отправку сообщений (без PII).
- Поддерживать корреляцию (`trace_id`, `request_id`, `message_id`, если есть).

## Ссылки

- Ошибки: `docs/design-guidelines/go/error_handling.md`.
