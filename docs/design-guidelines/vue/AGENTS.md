# Vue Design Guidelines

Документы для frontend (Vue 3 + TypeScript).

- `docs/design-guidelines/vue/check_list.md` — чек-лист перед PR для Vue изменений.
- `docs/design-guidelines/vue/frontend_architecture.md` — размещение, структура приложения, границы ответственности.
- `docs/design-guidelines/vue/frontend_data_and_state.md` — axios/Pinia/router/i18n/cookies/PWA/WebSocket правила.
- `docs/design-guidelines/vue/frontend_error_handling.md` — модель и обработка ошибок на фронте.
- `docs/design-guidelines/vue/frontend_code_rules.md` — правила кодирования (TS/Vue/импорты/комментарии).
- `docs/design-guidelines/vue/libraries.md` — что выносить в `libs/{vue,ts,js}` и как.
- `docs/design-guidelines/visual/AGENTS.md` — визуальный дизайн и UI‑паттерны.

Специфика `codex-k8s`:
- основной frontend приложения живет в `services/staff/web-console`;
- вход в UI защищен GitHub OAuth;
- пользовательские настройки и права приходят из backend API (PostgreSQL как source of truth).
- Планирование и закрытие daily задач в спринте — по
  `docs/delivery/development_process_requirements.md`.
