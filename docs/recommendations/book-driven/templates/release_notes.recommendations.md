# Рекомендации по доработке `docs/templates/release_notes.md`

- Дорабатываемый документ: [docs/templates/release_notes.md#L1-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/release_notes.md#L1-L60)
- Профиль рекомендаций: `template_delivery`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Зафиксировать release scope и критерии изменения границ
Что доработать: Добавить разделы «что входит/не входит в релиз» и «как обрабатываются новые требования после freeze».

Как внедрить: Для каждого изменения после freeze требовать экспресс-анализ влияния на сроки и риски.

Ожидаемый эффект: Это защищает релиз от неконтролируемого расширения объёма работ.

Выдержка из книги: [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610); [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798)

Фрагмент документа для изменения: [docs/templates/release_notes.md#L1-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/release_notes.md#L1-L60) (ориентир: `Документ целиком`).

### 2. Добавить явные связи релизного плана с требованиями и тестами
Что доработать: Включить таблицу: Requirement ID -> релизный инкремент -> критерий готовности -> подтверждающий тест.

Как внедрить: Обновлять таблицу при каждом переносе/добавлении требования.

Ожидаемый эффект: Улучшается контроль completeness и прогнозируемость поставки.

Выдержка из книги: [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062); [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311)

Фрагмент документа для изменения: [docs/templates/release_notes.md#L22-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/release_notes.md#L22-L60) (ориентир: `Release Notes: <Система> <vX.Y.Z>`).

### 3. Сделать этапы go/no-go измеримыми
Что доработать: Для решения о выпуске добавить количественные пороги по quality, стабильности и операционной готовности.

Как внедрить: Привязать пороги к заранее утверждённым acceptance-критериям.

Ожидаемый эффект: Решения о выпуске становятся повторяемыми и не зависят от субъективных оценок.

Выдержка из книги: [docs/source/book.md#L22946-L23122](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L22946-L23122); [docs/source/book.md#L18065-L19261](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L18065-L19261); [docs/source/book.md#L2924-L2968](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2924-L2968)

Фрагмент документа для изменения: [docs/templates/release_notes.md#L22-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/release_notes.md#L22-L60) (ориентир: `Release Notes: <Система> <vX.Y.Z>`).

### 4. Встроить атрибуты статуса и ответственных
Что доработать: Для всех ключевых задач/инкрементов хранить статус, владельца, дату изменения и целевой этап.

Как внедрить: Использовать эти атрибуты как обязательный формат отчётности по релизу.

Ожидаемый эффект: Пропадает «синдром 90% готовности», повышается точность планирования.

Выдержка из книги: [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109); [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599)

Фрагмент документа для изменения: [docs/templates/release_notes.md#L1-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/release_notes.md#L1-L60) (ориентир: `Документ целиком`).

### 5. Добавить управляемый процесс изменений и отката
Что доработать: В шаблон включить минимальный workflow изменения релизного плана и критерии запуска rollback.

Как внедрить: Фиксировать решение CCB/ответственных и ссылку на impact-анализ.

Ожидаемый эффект: Снижается риск хаотичных действий в критический период релиза.

Выдержка из книги: [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033); [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355)

Фрагмент документа для изменения: [docs/templates/release_notes.md#L22-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/release_notes.md#L22-L60) (ориентир: `Release Notes: <Система> <vX.Y.Z>`).

### 6. Ввести риск-реестр релиза с приоритизацией mitigation
Что доработать: Добавить обязательную секцию рисков с форматом причина -> следствие -> детектор -> mitigation -> владелец.

Как внедрить: Ранжировать риски по бизнес-влиянию и вероятности.

Ожидаемый эффект: Команда заранее готовит меры реагирования и избегает аврального управления релизом.

Выдержка из книги: [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798); [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510)

Фрагмент документа для изменения: [docs/templates/release_notes.md#L1-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/release_notes.md#L1-L60) (ориентир: `Документ целиком`).
