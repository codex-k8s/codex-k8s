# Рекомендации по доработке `docs/templates/docset_issue.md`

- Дорабатываемый документ: [docs/templates/docset_issue.md#L1-L49](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/docset_issue.md#L1-L49)
- Профиль рекомендаций: `template_doc_governance`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Сделать навигацию по документам ориентированной на аудитории
Что доработать: Добавить в шаблон явное разделение: какие документы для PM/SA/Dev/QA/SRE и в каком порядке их читать.

Как внедрить: Для каждого раздела указать обязательный минимальный набор артефактов.

Ожидаемый эффект: Снижается время поиска контекста и риск пропустить критичный документ.

Выдержка из книги: [docs/source/book.md#L12254-L12670](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L12254-L12670); [docs/source/book.md#L2416-L2830](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2416-L2830); [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510)

Фрагмент документа для изменения: [docs/templates/docset_issue.md#L19-L49](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/docset_issue.md#L19-L49) (ориентир: `DocSet: Issue #<N> — <Короткое название>`).

### 2. Добавить обязательную трассируемость между issue, документами и реализацией
Что доработать: В шаблонах фиксировать ссылки по цепочке: проблема -> требования -> архитектурное решение -> реализация -> тесты.

Как внедрить: Проверять целостность цепочки как часть review процесса.

Ожидаемый эффект: Повышается управляемость изменений и воспроизводимость решений по истории проекта.

Выдержка из книги: [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311); [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599); [docs/source/book.md#L24600-L24943](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24600-L24943)

Фрагмент документа для изменения: [docs/templates/docset_issue.md#L19-L49](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/docset_issue.md#L19-L49) (ориентир: `DocSet: Issue #<N> — <Короткое название>`).

### 3. Ввести quality gates для текстов и структуры документов
Что доработать: Добавить checklist, который проверяет полноту, недвусмысленность, актуальность и наличие критериев приемки.

Как внедрить: Для типовых ошибок добавить короткие анти-примеры.

Ожидаемый эффект: Качество документации станет стабильно высоким и менее зависимым от автора.

Выдержка из книги: [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574); [docs/source/book.md#L13575-L14319](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13575-L14319); [docs/source/book.md#L21983-L22604](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21983-L22604)

Фрагмент документа для изменения: [docs/templates/docset_issue.md#L1-L49](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/docset_issue.md#L1-L49) (ориентир: `Документ целиком`).

### 4. Добавить lifecycle-атрибуты документа
Что доработать: В каждом шаблоне предусмотреть: версия, статус, владелец, дата следующего пересмотра и степень актуальности.

Как внедрить: Эти поля использовать для построения реестра живых/устаревших документов.

Ожидаемый эффект: Проще поддерживать документацию в актуальном состоянии и планировать обновления.

Выдержка из книги: [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109); [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510)

Фрагмент документа для изменения: [docs/templates/docset_issue.md#L1-L49](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/docset_issue.md#L1-L49) (ориентир: `Документ целиком`).

### 5. Описать формальный процесс change request для документации
Что доработать: В шаблоны добавить минимальные требования к изменению: причина, влияние, риск, план верификации.

Как внедрить: Фиксировать результат рассмотрения и ссылку на принятое решение.

Ожидаемый эффект: Документация начнёт развиваться как управляемая система, а не набор разрозненных файлов.

Выдержка из книги: [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033); [docs/source/book.md#L2924-L2968](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2924-L2968)

Фрагмент документа для изменения: [docs/templates/docset_issue.md#L19-L49](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/docset_issue.md#L19-L49) (ориентир: `DocSet: Issue #<N> — <Короткое название>`).

### 6. Добавить риск-модель для документационного долга
Что доработать: Выделить риски устаревания: отсутствие владельца, отсутствующие ссылки, неактуальные шаблоны, разрывы в цепочке требований.

Как внедрить: Добавить меры профилактики и периодические ревизии.

Ожидаемый эффект: Снижается вероятность ошибок внедрения из-за неактуальной документации.

Выдержка из книги: [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510); [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798)

Фрагмент документа для изменения: [docs/templates/docset_issue.md#L19-L49](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/docset_issue.md#L19-L49) (ориентир: `DocSet: Issue #<N> — <Короткое название>`).
