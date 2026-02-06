# Design Guidelines

Документация разделена по областям:

- `docs/design-guidelines/common/` — общее для backend и frontend.
- `docs/design-guidelines/go/` — правила для Go backend.
- `docs/design-guidelines/vue/` — правила для frontend (Vue 3 + TypeScript).
- `docs/design-guidelines/visual/` — визуальный дизайн и UI‑паттерны.

Стартовая точка перед PR:
- `docs/design-guidelines/common/check_list.md`
- затем профильные чек-листы:
  - `docs/design-guidelines/go/check_list.md`
  - `docs/design-guidelines/vue/check_list.md`

Специфика `codex-k8s`, которую нельзя нарушать:
- только Kubernetes как оркестратор;
- webhook-driven процессы (без workflow-first оркестрации);
- PostgreSQL + JSONB + pgvector как источник синхронизации состояния;
- встроенные MCP сервисные ручки реализуются в Go внутри платформы;
- интеграции с репозиториями проектируются через интерфейсы провайдеров.
