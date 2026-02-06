# Рекомендации по доработке `docs/templates/definition_of_done.md`

- Дорабатываемый документ: [docs/templates/definition_of_done.md#L1-L62](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/definition_of_done.md#L1-L62)
- Профиль рекомендаций: `template_quality`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Связать тестовые артефакты с идентификаторами требований
Что доработать: Сделать обязательной колонку/поле Requirement ID для каждого тестового набора и тест-кейса.

Как внедрить: Поддерживать покрытие «требование -> тест -> результат» в актуальном виде на каждый релиз.

Ожидаемый эффект: Улучшается прозрачность покрытия и проще доказывать готовность к выпуску.

Выдержка из книги: [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062); [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311); [docs/source/book.md#L22637-L22945](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L22637-L22945)

Фрагмент документа для изменения: [docs/templates/definition_of_done.md#L1-L62](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/definition_of_done.md#L1-L62) (ориентир: `Документ целиком`).

### 2. Добавить критерии качества формулировки тестируемых требований
Что доработать: Перед включением в план проверять требования на полноту, недвусмысленность и проверяемость.

Как внедрить: Использовать чек-лист дефектов требований как входной quality gate для QA.

Ожидаемый эффект: Снижается доля «непроверяемых» требований и ложнопозитивных результатов тестирования.

Выдержка из книги: [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574); [docs/source/book.md#L13575-L14319](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13575-L14319); [docs/source/book.md#L21983-L22604](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21983-L22604)

Фрагмент документа для изменения: [docs/templates/definition_of_done.md#L1-L62](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/definition_of_done.md#L1-L62) (ориентир: `Документ целиком`).

### 3. Ввести риск-ориентированную приоритизацию тестов
Что доработать: Разделить тестовые сценарии на критичность по формуле value/cost/risk и зафиксировать минимальный набор для релиза.

Как внедрить: В релизной фазе при нехватке времени в первую очередь исполнять high-risk/high-value сценарии.

Ожидаемый эффект: Тестирование концентрируется на наиболее уязвимых областях продукта.

Выдержка из книги: [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798); [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610)

Фрагмент документа для изменения: [docs/templates/definition_of_done.md#L1-L62](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/definition_of_done.md#L1-L62) (ориентир: `Документ целиком`).

### 4. Добавить атрибуты состояния тестовой готовности
Что доработать: Для каждого блока тестов вести статус (planned/in-progress/passed/blocked/deferred), владельца и целевой релиз.

Как внедрить: Состояние публиковать как часть регулярной отчётности по quality gates.

Ожидаемый эффект: Менеджмент получает объективную картину зрелости тестирования вместо оценок «на глаз».

Выдержка из книги: [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109); [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599)

Фрагмент документа для изменения: [docs/templates/definition_of_done.md#L59-L62](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/definition_of_done.md#L59-L62) (ориентир: `Апрув`).

### 5. Описать change-процесс для тестовой документации
Что доработать: В шаблон добавить процедуру обновления тестовых артефактов при изменении требований/контрактов.

Как внедрить: Фиксировать, какие изменения требуют полного пересмотра матрицы тестов, а какие локального апдейта.

Ожидаемый эффект: Это предотвращает устаревание тестовой базы после частых изменений требований.

Выдержка из книги: [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062)

Фрагмент документа для изменения: [docs/templates/definition_of_done.md#L59-L62](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/definition_of_done.md#L59-L62) (ориентир: `Апрув`).

### 6. Усилить приемочные критерии и критерии остановки/возобновления
Что доработать: Определить количественные пороги для pass/fail, suspension/resumption и условия допуска в релиз.

Как внедрить: Связать эти пороги с бизнес-риском и договорённостями со стейкхолдерами.

Ожидаемый эффект: Команда сможет принимать решения по релизу на объективной и заранее согласованной основе.

Выдержка из книги: [docs/source/book.md#L22946-L23122](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L22946-L23122); [docs/source/book.md#L22637-L22945](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L22637-L22945); [docs/source/book.md#L2924-L2968](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2924-L2968)

Фрагмент документа для изменения: [docs/templates/definition_of_done.md#L1-L62](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/definition_of_done.md#L1-L62) (ориентир: `Документ целиком`).
