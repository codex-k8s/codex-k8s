# Рекомендации по доработке `docs/templates/problem.md`

- Дорабатываемый документ: [docs/templates/problem.md#L1-L48](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/problem.md#L1-L48)
- Профиль рекомендаций: `template_product`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Усилить раздел границ, целей и ограничений
Что доработать: Явно добавить в шаблон блоки: «входит в scope», «не входит в scope», «критерий завершения по границам».

Как внедрить: В каждый блок включить бизнес-обоснование и влияние на соседние домены.

Ожидаемый эффект: Команда получает устойчивый каркас против расползания требований в реализации.

Выдержка из книги: [docs/source/book.md#L5401-L6092](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L5401-L6092); [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610); [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599)

Фрагмент документа для изменения: [docs/templates/problem.md#L17-L48](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/problem.md#L17-L48) (ориентир: `Problem Statement: <Название>`).

### 2. Сделать представление пользователей и заинтересованных лиц обязательным
Что доработать: Добавить обязательные поля: классы пользователей, представитель класса, уровень привилегий, ключевые сценарии.

Как внедрить: Если ролей много, требовать ссылку на отдельный актуальный документ персон и матрицу ответственности.

Ожидаемый эффект: Требования будут отражать реальные сегменты пользователей, а не усреднённого «одного пользователя».

Выдержка из книги: [docs/source/book.md#L6799-L7202](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6799-L7202); [docs/source/book.md#L2416-L2830](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2416-L2830); [docs/source/book.md#L2924-L2968](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2924-L2968)

Фрагмент документа для изменения: [docs/templates/problem.md#L1-L48](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/problem.md#L1-L48) (ориентир: `Документ целиком`).

### 3. Добавить доказуемую проверяемость требований и критериев приемки
Что доработать: Каждое требование связывать с минимум одним критерией приемки и способом проверки (тест/демо/метрика).

Как внедрить: Ввести обязательный тест на недвусмысленность формулировки и полноту контекста требования.

Ожидаемый эффект: Документ станет готовой основой для QA и приемки, а не только описанием желаемого поведения.

Выдержка из книги: [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574); [docs/source/book.md#L13575-L14319](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13575-L14319); [docs/source/book.md#L22946-L23122](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L22946-L23122)

Фрагмент документа для изменения: [docs/templates/problem.md#L1-L48](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/problem.md#L1-L48) (ориентир: `Документ целиком`).

### 4. Добавить приоритизацию на основе ценности, стоимости и риска
Что доработать: В шаблон включить поля для оценки бизнес-ценности, стоимости реализации и риска/неопределённости по каждому крупному требованию.

Как внедрить: На основании этих полей формировать порядок реализации и границы MVP.

Ожидаемый эффект: Решения о приоритетах станут прозрачными и устойчивыми к субъективным спорам.

Выдержка из книги: [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798); [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610); [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355)

Фрагмент документа для изменения: [docs/templates/problem.md#L1-L48](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/problem.md#L1-L48) (ориентир: `Документ целиком`).

### 5. Добавить трассируемость документа к дизайну, коду и тестам
Что доработать: Предусмотреть обязательные ссылки: требование -> design doc/ADR -> тестовые артефакты -> эксплуатационные метрики.

Как внедрить: Для каждого ID требования добавить поле текущего статуса и версии.

Ожидаемый эффект: Упрощается контроль реализации, регрессий и выпуска по фактической готовности.

Выдержка из книги: [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311); [docs/source/book.md#L24600-L24943](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24600-L24943); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062)

Фрагмент документа для изменения: [docs/templates/problem.md#L1-L48](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/problem.md#L1-L48) (ориентир: `Документ целиком`).

### 6. Встроить политику изменений и риск-реестр для требований
Что доработать: Добавить разделы: процесс change request, критерии принятия/отклонения, журнал решений и карта рисков.

Как внедрить: Требовать анализ влияния для каждого изменения после утверждения baseline.

Ожидаемый эффект: Шаблон начнёт поддерживать управляемую эволюцию требований на всём цикле проекта.

Выдержка из книги: [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033); [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355)

Фрагмент документа для изменения: [docs/templates/problem.md#L45-L48](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/problem.md#L45-L48) (ориентир: `Апрув`).
