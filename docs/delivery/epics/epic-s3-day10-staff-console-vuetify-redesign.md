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
- Дополнительно: в навигации сразу закладываем `Admin / Cluster`, `Agents` и `System settings (locales)`.
- Обратная связь пользователю: единый паттерн `VAlert` (с иконками) + `VSnackbar` после действий (например, “XXX удален”).

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
  - группировка разделов по смыслу (Operations / Platform / Governance / Admin / Configuration).
- Навигация “в будущее”:
  - будущие разделы присутствуют в навигации и открываются как страницы (router + i18n);
  - будущие страницы наполнены Vuetify-компонентами и mock-данными;
  - в коде каждой заглушки есть `TODO` с конкретикой “что подключить дальше”).
- Административный блок `Admin / Cluster` (scaffold) для управления Kubernetes ресурсами (CRUD на уровне UI-заготовок + TODO на backend):
  - `Namespaces`;
  - `ConfigMaps`;
  - `Secrets`;
  - `Deployments`;
  - `Pods` + логи контейнеров;
  - `Jobs` + логи контейнеров;
  - `PVC`.
- Управление агентами (scaffold):
  - список агентов (system + custom, с разделением по проекту при необходимости);
  - настройки агента (execution mode, лимиты, policy-параметры);
  - шаблоны промптов агента (`work/review`) минимум в двух локалях (`ru` и `en`) с UI для переключения локали.
- System settings (scaffold):
  - управление локалями системы (добавление locale + выбор default locale);
  - TODO на интеграцию с backend-конфигом/БД.
- Production-ready UX-паттерны на Vuetify (не “черновая верстка”):
  - карточки/метрики: `VCard`;
  - списки/меню: `VList`, `VListItem`, `VMenu`;
  - статусы/бейджи: `VChip`, `VBadge`;
  - фильтры/поиск: `VTextField`, `VSelect` (плюс chips по месту);
  - таблицы/пагинация: `VDataTable` (или server-side variant) + `VPagination`;
  - диалоги подтверждения: `VDialog`;
  - загрузки/пустые состояния: `VSkeletonLoader` + общий empty-state (или `VEmptyState`, если доступен в выбранной версии);
  - обратная связь: `VAlert` (info/success/warning/error, с иконками) + `VSnackbar` после действий;
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
- Реальная backend-интеграция cluster CRUD (k8s API + RBAC + audit) для новых `Admin / Cluster` экранов (Day10 делает только UI scaffold + TODO).

## Ограничения безопасности для `Admin / Cluster` разделов (обязательны при дальнейшей реализации)
- Элементы платформы помечаем label `app.kubernetes.io/part-of=codex-k8s` (канонический критерий для UI и backend).
- Удаление элементов платформы запрещено.
- Элементы платформы в окружениях `ai-staging` и `prod` (namespaces вида `{{ .Project }}-ai-staging` и `{{ .Project }}-prod`) доступны только на просмотр (view-only):
  - скрывать/выключать действия create/update/delete;
  - показывать явный read-only banner на экранах.
- Для `ai` окружений (ai-slots; namespaces вида `{{ .Project }}-dev-{{ .Slot }}`) destructive действия должны отрабатывать на backend как dry-run:
  - кнопки действий в UI существуют (для dogfooding/debug), но реальный delete/apply не выполняется;
  - по клику пользователь получает обратную связь: “dry-run OK, но в этом режиме действие запрещено”.
- Для non-platform ресурсов (не помеченных `app.kubernetes.io/part-of=codex-k8s`) CRUD допускается, но:
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
  - `System settings` (mock + TODO): locales (add locale + default locale)
  - `Docs/knowledge` (mock + TODO)
  - `MCP tools: secret sync` (mock + TODO)
  - `MCP tools: databases` (mock + TODO)

## Требования к заглушкам (обязательны)
- Заглушка = не “пустая страница с текстом”, а минимальный UI-скелет:
  - `PageHeader` (заголовок + короткий hint);
  - `VCard`-метрики (2–4) или summary-карточка;
  - `FiltersBar` (по месту) с 1–3 контролами (`VTextField`, `VSelect`, `VChip`-filters);
  - `VDataTable`/`VList` с 5–15 строками mock-данных;
  - empty-state и loading-state (через общий компонент/слоты).
- Для экранов `Admin / Cluster` дополнительно:
  - selector namespace + баннер для view-only окружений `ai-staging`/`prod`;
  - действия create/edit/delete в view-only режиме скрыты/disabled;
  - для `ai` env destructive действия должны дергать backend dry-run (кнопка есть, действие не выполняется; показываем что-то вроде “dry-run OK, но изменить/удалить нельзя”);
  - для `Secret` mock показывать только metadata, значение скрыто/замазано.
- Обратная связь (обязательная база):
  - ошибки/предупреждения показываются через `VAlert` (с иконками);
  - после успешных действий показывается `VSnackbar` (например, “Deployment XXX удален”/“Сохранено”/“Обновлено”).
- В коде каждой заглушки должен быть `TODO`:
  - что именно подключить (store/api, endpoint, модель данных),
  - где ожидается контракт (OpenAPI endpoint, feature store).

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
- Story-4.1: Добавить scaffold “Admin / Cluster” (без данных):
  - экран/маршруты под каждый ресурс (namespaces/configmaps/secrets/deployments/pods/jobs/pvc);
  - детали ресурса (metadata, labels/annotations, status) + вкладка YAML/JSON view;
  - заготовки create/edit/delete flows (формы или YAML view) с disabled apply и `TODO: подключить staff API + audit`;
  - правила безопасности в TODO:
    - platform elements определяются по `app.kubernetes.io/part-of=codex-k8s`;
    - `ai-staging`/`prod` = view-only;
    - `ai` env = dry-run для destructive actions (кнопки есть, действие не применяется, есть feedback).
- Story-4.2: Добавить scaffold “Agents” и “System settings”:
  - agents list + agent details (tabs: Settings / Prompt templates / Audit);
  - prompt templates для `work/review` минимум в `ru/en` с переключателем локали;
  - system settings: locales table + add-locale dialog (mock) + `TODO: backend settings`.
- Story-5: Базовые переиспользуемые UI-компоненты (shared/ui):
  - `AppShell`, `AppDrawer`, `AppBarActions`;
  - `PageHeader`, `KpiCards`, `FiltersBar`;
  - `EmptyState`, `LoadingState` (skeleton presets);
  - `JsonViewer`/`CodeBlock` (для логов/JSON, с copy action);
  - `Notifications`: `VAlert` presets + `VSnackbar` helper для “action completed”.

## Критерии приемки
- В `services/staff/web-console` используется Vuetify app-shell:
  - navbar на `VAppBar`;
  - drawer на `VNavigationDrawer` (desktop + mobile поведение);
  - основной контент в `VMain`.
- В drawer присутствуют будущие разделы (scaffold) и они открываются как страницы (router работает).
- Scaffold-страницы содержат UI-скелет на компонентах Vuetify + mock-данные и `TODO: ...` о том, как их довести до production.
- В навигации присутствует `Admin / Cluster`, и для каждого ресурса есть страница scaffold.
- Для `ai-staging`/`prod` страницы `Admin / Cluster` явно показывают режим “только просмотр” и не предлагают destructive actions.
- Для `ai` окружений destructive действия в `Admin / Cluster` отрабатывают как dry-run и дают явную обратную связь “dry-run OK, но действие запрещено”.
- В навигации присутствуют `Agents` и `System settings (locales)` с UI-заготовками под:
  - настройки агента;
  - шаблоны промптов `work/review` минимум в локалях `ru/en`;
  - добавление locale в системе (mock + TODO).
- Текущие MVP-сценарии не регресснули:
  - runs list/details, jobs, waits, approvals, logs доступны и работают.
- На ключевых экранах использованы базовые Vuetify-компоненты (не ad-hoc HTML):
  - `VCard`, `VList`, `VChip`/`VBadge`, `VTextField`/`VSelect`, `VDataTable`, `VPagination`, `VMenu`, `VBtn`, `VIcon`, `VDialog`, `VSkeletonLoader`, `VAlert`, `VSnackbar`.
- Обратная связь пользователю реализована единообразно:
  - ошибки/предупреждения = `VAlert` (с иконками);
  - успешные действия = `VSnackbar` (например, “XXX удален”/“Сохранено”).
- Локализация: ключи меню/заголовков/пустых состояний покрыты `ru/en`.
- UI соответствует visual-гайдам:
  - светлая тема по умолчанию, спокойные поверхности, предсказуемая навигация,
  - корректные empty/loading/error состояния.
