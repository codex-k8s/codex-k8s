# Common Design Guidelines

Документы, общие для backend и frontend (не дублируются в `go/` и `vue/`).

- `docs/design-guidelines/common/check_list.md` — общий чек-лист (дальше — профильные в `go/` и `vue/`).
- `docs/design-guidelines/common/project_architecture.md` — зоны, границы ответственности, структура репо.
- `docs/design-guidelines/common/design_principles.md` — DDD/SOLID/DRY/KISS/Clean Architecture.
- `docs/design-guidelines/common/libraries_reusable_code_requirements.md` — общие правила выноса кода в `libs/*`.

Дополнительно для `codex-k8s`:
- процессы выполняются по webhook-событиям, а не через GitHub Actions workflows;
- Kubernetes и repository-провайдеры подключаются только через интерфейсы и адаптеры;
- модель данных и синхронизация multi-pod держатся на PostgreSQL (`JSONB` + `pgvector`).
