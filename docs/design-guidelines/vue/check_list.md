# Vue чек-лист перед PR

## Стек и размещение
- Используется стек: Vue 3 + TypeScript + Pinia + Axios + Vite + `vue-router` + `vue-i18n` + `vue3-cookies`.
- Приложение лежит в правильной зоне: `services/staff/*` (основной случай), `services/external/*` или `services/dev/*`.
- Перед написанием кода перечитаны:
  - `docs/design-guidelines/vue/frontend_architecture.md`
  - `docs/design-guidelines/vue/frontend_code_rules.md`
  - `docs/design-guidelines/vue/frontend_data_and_state.md`

## Структура и зависимости
- Структура `src/` выдержана (см. `docs/design-guidelines/vue/frontend_architecture.md`).
- Нет циклических зависимостей; `shared/*` не импортирует `features/pages`.

## Типы, модели, константы, helpers
- API DTO и transport-типы не объявлены ad-hoc в компонентах/страницах: используются `src/shared/api/generated/**`, `src/shared/api/*`, `src/features/*/types.ts`, `src/shared/types/*`.
- UI view-model отделён от API DTO: есть явный mapping/caster, нет смешивания “как пришло с API” и “как нужно UI”.
- Повторяющиеся литералы (route names, query keys, cookie/storage keys, status strings) вынесены в константы/enum/type-alias, нет копипасты строк по компонентам.
- Helper/composable-код размещён по уровню переиспользования:
  - локальный helper в компоненте только для одного места;
  - повторно используемый helper в `src/shared/lib/*` или `src/features/*/lib/*`;
  - межприложный код в `libs/{vue,ts}`.
- Нет ad-hoc `any`/unsafe cast для контрактов данных в components/stores; для невалидных внешних данных используется явный parser/validator.

## HTTP/API
- Нет прямых вызовов `axios` из компонентов/страниц; есть единый axios instance + интерсепторы + нормализация ошибок.
- UI не показывает сырые backend-ошибки; используются `messageKey`/i18n.
- Если менялись контракты API: `**/generated/**` обновлены через генерацию и не правились вручную.

## Auth и доступы
- Вход в UI соответствует GitHub OAuth flow.
- Guard'ы маршрутов опираются на серверный профиль пользователя и матрицу прав.
- Нет клиентских bypass-механизмов для прав доступа.

## State и UX
- Store по фичам; side-effects в actions/services; нет “глобального store на всё”.
- Пустые/ошибочные/загрузочные состояния оформлены по визуальным гайдам.

## libs
- Код, требуемый >= 2 приложениями, вынесен в `libs/{vue|ts}` по правилам общих гайдов.
- Перед пушем выполнена повторная сверка с `docs/design-guidelines/common/check_list.md` и текущим файлом; нарушения по размещению типов/моделей/констант/helper-кода устранены.
