# Рекомендации по доработке `docs/templates/risks_register.md`

- Дорабатываемый документ: [docs/templates/risks_register.md#L1-L37](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/risks_register.md#L1-L37)
- Профиль рекомендаций: `template_risk`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Стандартизовать запись рисков в формате «причина -> следствие»
Что доработать: Сделать этот формат обязательным полем шаблона и запретить абстрактные формулировки без причинности.

Как внедрить: Добавить примеры корректного и некорректного описания риска.

Ожидаемый эффект: Риски становятся пригодны для анализа и выбора адресных mitigation-мер.

Выдержка из книги: [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574); [docs/source/book.md#L13575-L14319](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13575-L14319)

Фрагмент документа для изменения: [docs/templates/risks_register.md#L17-L37](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/risks_register.md#L17-L37) (ориентир: `Risk Register: <Название>`).

### 2. Добавить связь риска с этапом работы с требованиями
Что доработать: Для каждого риска указывать, на каком этапе он проявляется: выявление, анализ, спецификация, утверждение, управление изменениями.

Как внедрить: Это позволяет назначать владельцев и контрольные точки по этапам.

Ожидаемый эффект: Повышается точность профилактики и распределение ответственности.

Выдержка из книги: [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L8572-L8660](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L8572-L8660); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033)

Фрагмент документа для изменения: [docs/templates/risks_register.md#L1-L37](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/risks_register.md#L1-L37) (ориентир: `Документ целиком`).

### 3. Ввести приоритизацию рисков по ценности, стоимости и вероятности
Что доработать: Добавить поля оценки для ранжирования mitigation и критерии, когда риск принимается осознанно.

Как внедрить: Фиксировать решение с обоснованием и сроком пересмотра.

Ожидаемый эффект: Команда концентрируется на рисках с максимальным бизнес-эффектом.

Выдержка из книги: [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798); [docs/source/book.md#L2924-L2968](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2924-L2968); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109)

Фрагмент документа для изменения: [docs/templates/risks_register.md#L1-L37](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/risks_register.md#L1-L37) (ориентир: `Документ целиком`).

### 4. Добавить трассируемость риска к требованиям и артефактам контроля
Что доработать: Для каждого риска указывать связанные требования, документы, тесты и операционные метрики.

Как внедрить: Поддерживать обновляемую матрицу связей в релизном цикле.

Ожидаемый эффект: Легче оценивать фактическое покрытие рисков и находить пробелы контроля.

Выдержка из книги: [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062); [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599)

Фрагмент документа для изменения: [docs/templates/risks_register.md#L17-L37](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/risks_register.md#L17-L37) (ориентир: `Risk Register: <Название>`).

### 5. Добавить lifecycle и статусное управление рисками
Что доработать: Вести статусы риска (new/assessed/mitigating/accepted/closed), дату ревизии и ответственного.

Как внедрить: Отражать статус в регулярной отчётности проекта.

Ожидаемый эффект: Снижается число «вечных» открытых рисков без решения.

Выдержка из книги: [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109); [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510)

Фрагмент документа для изменения: [docs/templates/risks_register.md#L17-L37](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/risks_register.md#L17-L37) (ориентир: `Risk Register: <Название>`).

### 6. Синхронизировать реестр рисков с процессом change control
Что доработать: Изменения с высоким риском автоматически переводить в formal change request с impact-анализом.

Как внедрить: Решения CCB и критерии отката делать обязательной частью записи риска.

Ожидаемый эффект: Повышается управляемость рисковых изменений и качество релизов.

Выдержка из книги: [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033); [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355)

Фрагмент документа для изменения: [docs/templates/risks_register.md#L1-L37](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/risks_register.md#L1-L37) (ориентир: `Документ целиком`).
