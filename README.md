# codex-k8s

`codex-k8s` — cloud-сервис (Go + Vue3), который объединяет функциональные требования
`codexctl`, `yaml-mcp-server` и `machine_driven_company_requirements` в единую
webhook-driven платформу для работы AI-агентов в Kubernetes.

## Базовые принципы

- только Kubernetes (через Go SDK),
- repository providers через интерфейсы (GitHub сейчас, GitLab позже),
- без workflow-first оркестрации: процессы запускаются webhook-событиями,
- PostgreSQL как центральный state backend (`JSONB` + `pgvector`),
- встроенные служебные MCP ручки в Go,
- staff frontend на Vue3, защищённый GitHub OAuth.

## Каркас репозитория

- `services/external/api-gateway` — внешний API/webhook ingress.
- `services/staff/web-console` — UI/настройки/наблюдение.
- `services/internal/control-plane` — доменная логика платформы.
- `services/jobs/worker` — фоновые jobs/reconciliation.
- `services/dev/webhook-simulator` — dev-only инструменты.
- `docs/design-guidelines` — обязательные инженерные стандарты.

## Стандарты разработки

Перед изменениями обязательно читать:
- `AGENTS.md`
- `docs/design-guidelines/AGENTS.md`
