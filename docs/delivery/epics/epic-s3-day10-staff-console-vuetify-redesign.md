---
doc_id: EPC-CK8S-S3-D10
type: epic
title: "Epic S3 Day 10: Staff console on Vuetify (new app-shell + navigation scaffold)"
status: planned
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-15
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 10: Staff console on Vuetify (new app-shell + navigation scaffold)

## TL;DR
- Цель: не “перекрасить” текущий UI, а заложить основу новой staff-консоли, которая сразу готова к post-MVP расширению.
- MVP-результат дня: production-ready `Vuetify` app-shell (navbar + drawer) + навигация по будущим разделам + единые UI-паттерны (таблицы/фильтры/пустые состояния/диалоги).
- Важно: будущие разделы не интегрируем с API/Pinia (без новых store и backend-запросов), но делаем страницы и компоненты с заглушками данных и явными `TODO`.

## Priority
- `P0`.

## Контекст
- Сейчас `services/staff/web-console` уже закрывает часть MVP-сценариев (projects/runs/runtime debug/approvals), но UI собран на ad-hoc CSS и разрозненных паттернах.
- По требованиям MVP staff-консоль должна масштабироваться до “операционного workspace” (runs, approvals, docs/templates, agents, labels/stages, audit). См.:
  - `docs/product/requirements_machine_driven.md` (FR-040..FR-046 + post-MVP направления),
  - `docs/product/brief.md` (post-MVP UI направления),
  - `docs/design-guidelines/visual/*` (навигация/таблицы/состояния),
  - `docs/design-guidelines/vue/*` (структура/границы/ошибки/state).

## Scope
### In scope
- Полная миграция staff-консоли на `Vuetify` (Vue 3) с корректной интеграцией:
  - зависимости и интеграция: `vuetify` + `vite-plugin-vuetify` + иконки (базовый вариант: `@mdi/font`);
  - app-shell: `VApp` + `VAppBar` (navbar) + `VNavigationDrawer` (drawer + rail) + `VMain`.
- Header/brand:
  - логотип `codex-k8s` (источник: `https://github.com/codex-k8s/codexctl/blob/5a0825435d9eaad9f9e52e745f9dcc5d683e59e6/docs/media/logo.png`);
  - favicon из того же источника (преобразовать в `.ico` при необходимости).
- Новая информационная архитектура:
  - левый drawer как primary-навигация;
  - группировка разделов по смыслу (Operations / Platform / Governance / Future).
- Навигация “в будущее”:
  - будущие разделы присутствуют в навигации и открываются как страницы (router + i18n);
  - будущие страницы наполнены Vuetify-компонентами и mock-данными;
  - в коде каждой заглушки есть `TODO` с конкретикой “что подключить дальше” и ссылкой на issue (минимум `#19`, если нет отдельного).
- Production-ready UX-паттерны на Vuetify (не “черновая верстка”):
  - карточки/метрики: `VCard`;
  - списки/меню: `VList`, `VListItem`, `VMenu`;
  - статусы/бейджи: `VChip`, `VBadge`;
  - фильтры/поиск: `VTextField`, `VSelect` (плюс chips по месту);
  - таблицы/пагинация: `VDataTable` (или server-side variant) + `VPagination`;
  - диалоги подтверждения: `VDialog`;
  - загрузки/пустые состояния: `VSkeletonLoader` + общий empty-state (или `VEmptyState`, если доступен в выбранной версии);
  - действия/иконки: `VBtn`, `VIcon`.
- UI/UX паритет для MVP-операционных сценариев (реальные данные остаются реальными):
  - runs list + run details;
  - running jobs / wait queue / approvals / run logs.
- UI governance:
  - i18n для меню/заголовков/пустых состояний (RU/EN);
  - соответствие визуальным гайдам (светлая тема по умолчанию, спокойные поверхности, предсказуемая навигация).

### Out of scope
- Реализация бизнес-логики будущих разделов:
  - без новых API endpoint,
  - без новых Pinia store,
  - без реальных данных из backend (кроме уже существующих MVP-экранов).
- Post-MVP функционал “по-настоящему”:
  - template editor 2.0,
  - agent constructor,
  - analytics studio,
  - governance UI для изменения stage/label policy.

## Целевая навигация (секции и статус наполнения)
Принцип: drawer показывает весь будущий “скелет” консоли, но многие разделы помечены как “coming soon” и работают на mock-данных.

Рекомендуемая карта разделов (минимум):
- Operations:
  - `Runs` (реально работает)
  - `Run details` (реально работает, переход из Runs)
  - `Running jobs` (реально работает; допускается вынести как отдельный экран)
  - `Wait queue` (реально работает; допускается вынести как отдельный экран)
  - `Approvals` (реально работает; допускается вынести как отдельный экран)
  - `Logs` (реально работает из run details; отдельный экран допускается как scaffold)
- Platform:
  - `Projects` (реально работает)
  - `Project details` (реально работает)
  - `Repositories` (реально работает)
  - `Members` (реально работает)
  - `Users` (реально работает, admin-only)
- Governance (scaffold):
  - `Audit log` (mock + TODO)
  - `Labels & stages` (mock + TODO)
- Future (scaffold):
  - `Agents` (mock + TODO)
  - `Prompt templates` (mock + TODO)
  - `Docs/knowledge` (mock + TODO)
  - `Control tools: secrets` (mock + TODO)
  - `Control tools: databases` (mock + TODO)

## Требования к заглушкам (обязательны)
- Заглушка = не “пустая страница с текстом”, а минимальный UI-скелет:
  - `PageHeader` (заголовок + короткий hint);
  - `VCard`-метрики (2–4) или summary-карточка;
  - `FiltersBar` (по месту) с 1–3 контролами (`VTextField`, `VSelect`, `VChip`-filters);
  - `VDataTable`/`VList` с 5–15 строками mock-данных;
  - empty-state и loading-state (через общий компонент/слоты).
- В коде каждой заглушки должен быть `TODO`:
  - что именно подключить (store/api, endpoint, модель данных),
  - где ожидается контракт (OpenAPI endpoint, feature store),

## Декомпозиция (Stories/Tasks)
- Story-0: Governance по зависимостям:
  - уточнить актуальные версии через Context7;
  - добавить `Vuetify` и иконки в `docs/design-guidelines/common/external_dependencies_catalog.md`.
- Story-1: Подключить Vuetify (Vite + Vue3):
  - добавить зависимости `vuetify`, `vite-plugin-vuetify`, `@mdi/font`;
  - настроить `createVuetify()` (icons + theme);
  - перевести composition root на `VApp` и убрать legacy layout-CSS, где он конфликтует.
- Story-2: Реализовать app-shell:
  - `VAppBar` (navbar): бренд/заголовок/crumbs + справа language switch + user menu/logout;
  - `VNavigationDrawer`: группы разделов, active state, responsive поведение (mobile temporary, desktop permanent/rail);
  - `VMain`: единый контейнер контента, предсказуемые отступы/ширины.
- Story-3: Привести существующие страницы к Vuetify-паттернам (с сохранением поведения):
  - Projects/ProjectDetails/Repos/Members/Users;
  - Runs/RunDetails/Approvals/Jobs/Waits/Logs (разнести или оставить, но UI должен стать компонентным и единообразным).
- Story-4: Добавить scaffold будущих разделов (без данных):
  - router paths + i18n keys + drawer links;
  - page-skeleton с mock-данными и `TODO`-метками.
- Story-5: Базовые переиспользуемые UI-компоненты (shared/ui):
  - `AppShell`, `AppDrawer`, `AppBarActions`;
  - `PageHeader`, `KpiCards`, `FiltersBar`;
  - `EmptyState`, `LoadingState` (skeleton presets);
  - `JsonViewer`/`CodeBlock` (для логов/JSON, с copy action).

## Критерии приемки
- В `services/staff/web-console` используется Vuetify app-shell:
  - navbar на `VAppBar`;
  - drawer на `VNavigationDrawer` (desktop + mobile поведение);
  - основной контент в `VMain`.
- В drawer присутствуют будущие разделы (scaffold) и они открываются как страницы (router работает).
- Future-страницы содержат UI-скелет на компонентах Vuetify + mock-данные и `TODO(#19): ...` (или более точный issue) о том, как их довести до production.
- Текущие MVP-сценарии не регресснули:
  - runs list/details, jobs, waits, approvals, logs доступны и работают.
- На ключевых экранах использованы базовые Vuetify-компоненты (не ad-hoc HTML):
  - `VCard`, `VList`, `VChip`/`VBadge`, `VTextField`/`VSelect`, `VDataTable`, `VPagination`, `VMenu`, `VBtn`, `VIcon`, `VDialog`, `VSkeletonLoader`.
- Локализация: ключи меню/заголовков/пустых состояний покрыты `ru/en`.
- UI соответствует visual-гайдам:
  - светлая тема по умолчанию, спокойные поверхности, предсказуемая навигация,
  - корректные empty/loading/error состояния.
