---
doc_id: EPC-CK8S-S3-D10
type: epic
title: "Epic S3 Day 10: Staff console on Vuetify (new app-shell + navigation scaffold)"
status: completed
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-16
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
- Сквозные элементы верстки (обязательная база): context selector, notifications menu, `VAlert`/`VSnackbar`, breadcrumbs + copy IDs, “coming soon” badges, table settings, row actions menu.
- Для редакторов markdown и YAML используем Monaco Editor (и только для них).

## Priority
- `P0`.

## Контекст
- Сейчас `services/staff/web-console` уже закрывает часть MVP-сценариев (projects/runs/runtime debug/approvals), но UI собран на ad-hoc CSS и разрозненных паттернах.
- По требованиям MVP staff-консоль должна масштабироваться до “операционного workspace” (runs, approvals, docs/templates, agents, labels/stages, audit, cluster debug). См.:
  - `docs/product/requirements_machine_driven.md` (FR-040..FR-046 + post-MVP направления),
  - `docs/product/brief.md` (post-MVP UI направления),
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
  - группировка разделов по смыслу (Operations / Platform / Governance / Admin / Configuration);
  - “coming soon” бейджи у scaffold-разделов.

- Сквозные UX/паттерны (обязательны в базовой верстке):
  - глобальный selector контекста в `VAppBar`: project, env, namespace/slot (даже если пока не влияет на данные);
  - notifications menu (иконка + `VMenu`) с лентой событий (mock + TODO);
  - единый слой обратной связи:
    - ошибки/предупреждения через `VAlert` разных типов с иконками;
    - после действий `VSnackbar` (например, “XXX удален”/“Сохранено”/“Обновлено”);
  - breadcrumbs + copy IDs в шапке страниц (run_id, correlation_id, namespace) с `VTooltip`;
  - таблицы: settings (density, column visibility, sticky header);
  - действия строки: через `VMenu` “⋯” (единый паттерн row actions);
  - унифицированные состояния: skeleton/empty/error с CTA “Повторить” и (при наличии) request-id.

- Operations (реальные данные остаются реальными, без регрессий):
  - master-detail layout для части сценариев (например, runs/approvals) вместо только отдельных страниц;
  - runs: stage timeline/stepper (показывает stage/status, ссылки на issue/PR/namespace, быстрые copy);
  - logs viewer компонент:
    - follow tail,
    - поиск по строкам,
    - подсветка уровней,
    - copy block,
    - download;
  - approvals center как отдельный экран (фильтры по tool/action/state, история решений) или переработка текущего `Approvals` блока до “центра”.

- Admin / Cluster (scaffold) для управления Kubernetes ресурсами (CRUD на уровне UI-заготовок + TODO на backend):
  - `Namespaces`;
  - `ConfigMaps`;
  - `Secrets`;
  - `Deployments`;
  - `Pods` + логи контейнеров;
  - `Jobs` + логи контейнеров;
  - `PVC`.
  - cluster-global selector namespace + mode banner (view-only/dry-run/normal);
  - страница ресурса с табами: Overview, YAML, Events, Related (и Logs для pod/job);
  - action preview перед destructive действиями;
  - `Secrets`: по умолчанию только metadata + список ключей, reveal значения как отдельное (будущее) действие.

- Управление агентами (scaffold):
  - список агентов с чипами: system/custom, mode (full-env/code-only), лимиты, статус;
  - карточка агента с табами: Settings / Prompt templates / History-Audit;
  - prompt templates:
    - переключатель locale `ru/en`,
    - diff/preview,
    - preview “effective template”;
    - редактор templates на Monaco Editor (markdown).

- System settings (scaffold):
  - таблица локалей: default locale, add locale dialog (mock + TODO);
  - глобальные UI-параметры (scaffold): density, формат дат/времени, debug hints.

- Governance (scaffold):
  - audit log экран с filter-bar (actor, object, env, correlation_id) + mock rows.

- Docs/knowledge (scaffold):
  - layout “левый сайдбар (дерево) + контент + правый TOC”;
  - code-blocks с copy;
  - markdown editor на Monaco Editor.

- MCP tools (scaffold):
  - каталог инструментов;
  - матрица апрувов;
  - история вызовов/outcome (mock).

- Production-ready UI-набор на Vuetify (не “черновая верстка”):
  - карточки/метрики: `VCard`;
  - списки/меню: `VList`, `VListItem`, `VMenu`;
  - статусы/бейджи: `VChip`, `VBadge`;
  - фильтры/поиск: `VTextField`, `VSelect`;
  - таблицы/пагинация: `VDataTable` (или server-side variant) + `VPagination`;
  - диалоги: `VDialog`;
  - состояния: `VSkeletonLoader` + общий empty-state.

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
- Реальная backend-интеграция cluster CRUD (k8s API + RBAC + audit) для новых `Admin / Cluster` экранов (Day10 делает только UI scaffold + TODO).

## Ограничения безопасности для `Admin / Cluster` разделов (обязательны при дальнейшей реализации)
- В окружениях `production` и `prod` элементы платформы помечаем label `app.kubernetes.io/part-of=codex-k8s` (канонический критерий для UI и backend).
- Исключение для dogfooding: в `ai` окружениях (ai-slots) платформа может разворачиваться без `app.kubernetes.io/part-of=codex-k8s`, чтобы UI позволял dogfood/debug действия над самой платформой (в т.ч. destructive через dry-run) и не применял platform guardrails по label.
- Ресурсы, помеченные `app.kubernetes.io/part-of=codex-k8s`, нельзя удалять.
- Элементы платформы в окружениях `production` и `prod` (namespaces вида `{{ .Project }}-production` и `{{ .Project }}-prod`) доступны только на просмотр (view-only):
  - скрывать/выключать действия create/update/delete;
  - показывать явный read-only banner на экранах.
- Для `ai` окружений (ai-slots; namespaces вида `{{ .Project }}-dev-{{ .Slot }}`) destructive действия должны отрабатывать на backend как dry-run:
  - кнопки действий в UI существуют (для dogfooding/debug), но реальный delete/apply не выполняется;
  - по клику пользователь получает обратную связь: “dry-run OK, но в этом режиме действие запрещено”.
- Для non-platform ресурсов (не помеченных `app.kubernetes.io/part-of=codex-k8s`; включая платформу в ai-slots при dogfooding) CRUD допускается, но:
  - destructive actions только после явного подтверждения (dialog) и с аудитом;
  - значения `Secret` по умолчанию не показывать (только метаданные); вывод значения и редактирование должны быть отдельным осознанным действием.

## Целевая навигация (секции и статус наполнения)
Принцип: drawer показывает весь будущий “скелет” консоли, но многие разделы помечены как “coming soon” и работают на mock-данных.

Рекомендуемая карта разделов (минимум):
- Operations:
  - `Runs` (реально работает)
  - `Run details` (реально работает, переход из Runs)
  - `Running jobs` (реально работает; допускается вынести как отдельный экран)
  - `Wait queue` (реально работает; допускается вынести как отдельный экран)
  - `Approvals` (реально работает; переработка до approvals center)
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
- Admin / Cluster (scaffold):
  - `Namespaces` (mock + TODO)
  - `ConfigMaps` (mock + TODO)
  - `Secrets` (mock + TODO)
  - `Deployments` (mock + TODO)
  - `Pods` (mock + TODO)
  - `Pod logs` (mock + TODO)
  - `Jobs` (mock + TODO)
  - `Job logs` (mock + TODO)
  - `PVC` (mock + TODO)
- Configuration (scaffold):
  - `Agents` (mock + TODO): list, details, settings, prompt templates (`ru/en`)
  - `System settings` (mock + TODO): locales + UI prefs
  - `Docs/knowledge` (mock + TODO)
  - `MCP tools` (mock + TODO)

## Требования к заглушкам (обязательны)
- Заглушка = не “пустая страница с текстом”, а минимальный UI-скелет:
  - `PageHeader` (заголовок + короткий hint);
  - `VCard`-метрики (2–4) или summary-карточка;
  - `FiltersBar` (по месту) с 1–3 контролами (`VTextField`, `VSelect`, `VChip`-filters);
  - `VDataTable`/`VList` с 5–15 строками mock-данных;
  - empty-state + loading-state + error-state.
- Табличные scaffold-экраны обязаны использовать общий toolbar:
  - table settings (density/columns) и row actions через `VMenu` “⋯”.
- Для экранов `Admin / Cluster` дополнительно:
  - selector namespace + баннер режима (view-only/dry-run);
  - действия create/edit/delete в view-only режиме скрыты/disabled;
  - для `ai` env destructive действия дергают backend dry-run (кнопка есть, действие не выполняется; показываем “dry-run OK, но тут удалять нельзя”);
  - `Secret` по умолчанию: metadata only.
- Обратная связь (обязательная база):
  - ошибки/предупреждения показываются через `VAlert` (с иконками);
  - после успешных действий показывается `VSnackbar`.
- Monaco Editor:
  - markdown editor (docs/prompt templates) = Monaco;
  - YAML editor/view (cluster resources) = Monaco;
  - для остальных code blocks использовать простой `CodeBlock`.
- В коде каждой заглушки должен быть `TODO`:
  - что именно подключить (store/api, endpoint, модель данных),
  - где ожидается контракт (OpenAPI endpoint, feature store),
  - ссылка на issue (`TODO(#19): ...` как минимум).

## Декомпозиция (Stories/Tasks)
- Story-0: Governance по зависимостям:
  - уточнить актуальные версии через Context7;
  - добавить `Vuetify`, иконки и Monaco Editor в `docs/design-guidelines/common/external_dependencies_catalog.md`.
- Story-1: Подключить Vuetify (Vite + Vue3):
  - добавить зависимости `vuetify`, `vite-plugin-vuetify`, `@mdi/font`;
  - настроить `createVuetify()` (icons + theme).
- Story-2: Реализовать app-shell:
  - `VAppBar`: бренд + breadcrumbs + context selector + notifications menu + user menu/logout;
  - `VNavigationDrawer`: группы разделов, active state, “coming soon” badges;
  - `VMain`: единый контейнер контента, предсказуемые отступы/ширины.
- Story-3: Shared UI foundation (shared/ui):
  - feedback layer: `VAlert` presets + очередь `VSnackbar`;
  - `BreadcrumbsBar` + copy IDs;
  - `DataTable` wrapper с table settings (density/columns) + row actions menu;
  - состояния `LoadingState`/`EmptyState`/`ErrorState`;
  - `MasterDetailLayout`.
- Story-3.1: Operations UX:
  - run timeline/stepper;
  - logs viewer компонент (tail/search/highlight/copy/download);
  - approvals center экран (или эквивалентная переработка существующего блока).
- Story-4: Привести существующие страницы к Vuetify-паттернам (с сохранением поведения):
  - Projects/ProjectDetails/Repos/Members/Users;
  - Runs/RunDetails/Approvals/Jobs/Waits/Logs.
- Story-4.1: Добавить scaffold “Admin / Cluster” (без данных):
  - экран/маршруты под каждый ресурс (namespaces/configmaps/secrets/deployments/pods/jobs/pvc);
  - детали ресурса с табами (Overview/YAML/Events/Related/Logs);
  - action preview диалог;
  - правила безопасности и режимов в TODO:
    - в `production`/`prod` platform elements определяются по `app.kubernetes.io/part-of=codex-k8s`; в ai-slots этот label может отсутствовать (dogfooding);
    - `production`/`prod` = view-only (для ресурсов с `app.kubernetes.io/part-of=codex-k8s`);
    - `ai` env = dry-run для destructive actions (кнопки есть, действие не применяется, есть feedback).
- Story-4.2: Добавить scaffold “Agents” и “System settings”:
  - agents list (чипы system/custom/mode/limits/status);
  - agent details tabs (Settings/Prompt templates/History-Audit);
  - prompt templates editor: locale `ru/en` + diff/preview + effective template preview (Monaco markdown);
  - system settings: locales table + add-locale dialog + UI prefs scaffold.
- Story-4.3: Добавить scaffold “Audit log / Docs / MCP tools”:
  - audit log: filter-bar + mock rows;
  - docs layout + markdown editor (Monaco);
  - MCP tools catalog + approval matrix + history (mock).
- Story-5: Monaco Editor integration:
  - добавить `monaco-editor`;
  - настроить Vite/worker интеграцию;
  - общий wrapper компонент под Monaco;
  - использовать Monaco только для markdown и YAML.

## Критерии приемки
- В `services/staff/web-console` используется Vuetify app-shell:
  - navbar на `VAppBar`;
  - drawer на `VNavigationDrawer` (desktop + mobile поведение);
  - основной контент в `VMain`.
- В app bar присутствуют: context selector, notifications menu, breadcrumbs.
- В drawer присутствуют будущие разделы (scaffold) и они открываются как страницы (router работает), включая “coming soon” badges.
- Scaffold-страницы содержат UI-скелет на компонентах Vuetify + mock-данные и `TODO(#19): ...`.
- Табличные экраны используют единый паттерн table settings + row actions menu.
- Обратная связь пользователю реализована единообразно:
  - ошибки/предупреждения = `VAlert` (с иконками);
  - успешные действия = `VSnackbar`.
- Operations UI расширен:
  - есть run timeline/stepper;
  - есть logs viewer с tail/search/highlight/copy/download;
  - approvals представлены как “центр” (отдельный экран или эквивалентная переработка).
- В навигации присутствует `Admin / Cluster`, и для каждого ресурса есть страница scaffold.
- Для `production`/`prod` страницы `Admin / Cluster` показывают режим “только просмотр”.
- Для `ai` окружений destructive действия в `Admin / Cluster` отрабатывают как dry-run и дают явную обратную связь “dry-run OK, но действие запрещено”.
- Есть `Agents` и `System settings (locales)` scaffolds:
  - prompt templates `work/review` минимум в `ru/en` с diff/preview + effective template preview;
  - locales table + add locale dialog;
  - UI prefs scaffold.
- Есть scaffolds `Audit log`, `Docs/knowledge`, `MCP tools`.
- Monaco Editor используется только для markdown и YAML редакторов.
- Текущие MVP-сценарии не регресснули:
  - runs list/details, jobs, waits, approvals, logs доступны и работают.
- Локализация: ключи меню/заголовков/пустых состояний покрыты `ru/en`.
