# Frontend: архитектура и размещение

## Где живёт фронтенд

- Основная консоль `codex-k8s` размещается в `services/staff/web-console`.
- Дополнительные публичные страницы (если появятся) размещаются в `services/external/*`.
- Dev-only frontend размещается в `services/dev/*` и не деплоится в production.

## Container и deploy требования

- Каждый frontend-сервис обязан иметь `Dockerfile` в корне сервиса.
- В Dockerfile обязательно поддерживаются target:
  - `dev` — для production/dev (например, Vite dev server);
  - `prod` — для production runtime со статическим бандлом (например, `nginx`).
- Для каждого frontend-сервиса обязателен отдельный Kubernetes manifest-шаблон
  в `deploy/base/<service>/*.yaml.tpl`.

## Технологический стек

- Vue 3
- TypeScript
- Pinia (state)
- Axios (HTTP)
- Vite + `@vitejs/plugin-vue` + `vite-plugin-pwa`
- `vue-i18n`
- `vue-router`
- `vue3-cookies`

## Границы ответственности

- Frontend не реализует backend бизнес-логику.
- Контракт с backend: HTTP/OpenAPI (`api/server/api.yaml`) и стабильный формат ошибок.
- Контракт async сообщений (если есть): AsyncAPI (`api/server/asyncapi.yaml`).
- OAuth state и permissions приходят из backend; frontend не хранит доверенные полномочия локально как источник истины.

## Рекомендуемая структура приложения

Внутри `services/staff/web-console/`:
- `index.html`, `package.json`, `vite.config.ts`.
- `public/` — статические файлы.
- `src/`:
  - `src/app/` — composition root: app/router/pinia/i18n/PWA.
  - `src/router/` — маршруты и guards.
  - `src/i18n/` — словари и ключи.
  - `src/shared/`:
    - `src/shared/api/` — axios client и нормализация ошибок.
    - `src/shared/api/generated/` — codegen артефакты.
    - `src/shared/ws/` — WS клиент (если используется).
    - `src/shared/ui/` — базовые UI-компоненты.
    - `src/shared/lib/` — утилиты.
  - `src/features/` — фичи и Pinia stores.
  - `src/pages/` — route-level страницы.

## Запрещено

- Прямые вызовы `axios` из компонентов/страниц (только через слой `shared/api`/feature services).
- Хранить секреты/token material в локальном хранилище как источник истины.
- Дублировать права доступа в UI без серверной валидации.
