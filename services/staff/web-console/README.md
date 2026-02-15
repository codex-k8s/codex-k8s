# web-console

`web-console` — staff frontend (Vue 3 + TypeScript) для управления пользователями, проектами, репозиториями и запуском/диагностикой agent run.

```text
services/staff/web-console/                          staff UI-приложение платформы
├── README.md                                        карта структуры frontend-сервиса
├── Dockerfile                                       multi-target сборка (`dev`/`prod`) для staging и runtime
├── package.json                                     зависимости и npm-скрипты приложения
├── package-lock.json                                lockfile зависимостей Node.js
├── index.html                                       HTML-шаблон точки входа Vite
├── vite.config.ts                                   конфигурация Vite-сборки
├── tsconfig.json                                    TypeScript-конфигурация проекта
├── openapi-ts.config.ts                             генерация typed API-клиента из OpenAPI
├── docker/nginx/default.conf                        runtime-конфиг web-сервера для prod target
└── src/                                             исходный код приложения
    ├── main.ts                                      bootstrap приложения
    ├── app/                                         корневой App и глобальные стили
    ├── i18n/                                        локализация интерфейса
    ├── router/                                      маршрутизация страниц
    ├── pages/                                       page-level компоненты экранов
    ├── features/                                    feature-модули (auth/projects/runs/users)
    └── shared/                                      переиспользуемый слой UI/lib/api
        └── api/                                     typed API-клиенты и контракты транспорта
```

## ai-staging: Vite HMR за Ingress

В `ai-staging` UI работает в режиме `vite dev server` за HTTPS Ingress и reverse-proxy.
Чтобы Vite HMR не пытался подключаться к `localhost:5173` в браузере и websocket стабильно работал,
в деплоймент пробрасываются переменные:

- `VITE_ALLOWED_HOSTS` (публичный домен)
- `VITE_HMR_HOST` (публичный домен)
- `VITE_HMR_PROTOCOL=wss`
- `VITE_HMR_CLIENT_PORT=443`
- `VITE_HMR_PATH=/__vite_ws`
