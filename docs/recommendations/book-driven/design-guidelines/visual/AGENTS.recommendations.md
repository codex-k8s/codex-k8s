# Рекомендации по доработке `docs/design-guidelines/visual/AGENTS.md`

- Дорабатываемый документ: [docs/design-guidelines/visual/AGENTS.md#L1-L15](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/visual/AGENTS.md#L1-L15)
- Профиль рекомендаций: `guideline_agents`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Явно зафиксировать рамки действия инструкций и зоны исключений
Что доработать: Добавить в документ раздел, где перечислены: что покрывает инструкция, что вне области действия, и при каких условиях допустимы отступления.

Как внедрить: Описать формат исключения: инициатор, обоснование, срок действия, кто утверждает и где фиксируется решение.

Ожидаемый эффект: Снижается риск неявных трактовок и «ползучего» расширения правил; проще обучать новых участников проекта.

Выдержка из книги: [docs/source/book.md#L5401-L6092](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L5401-L6092); [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610); [docs/source/book.md#L12254-L12670](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L12254-L12670)

Фрагмент документа для изменения: [docs/design-guidelines/visual/AGENTS.md#L1-L15](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/visual/AGENTS.md#L1-L15) (ориентир: `Документ целиком`).

### 2. Добавить матрицу ролей и прав на принятие решений
Что доработать: Уточнить, кто формулирует требования к стандартам, кто согласует конфликтные случаи и кто отвечает за аудит исполнения.

Как внедрить: Оформить таблицу RACI для ролей (author/reviewer/approver) и прописать SLA по времени реакции на конфликтные изменения.

Ожидаемый эффект: Повышается предсказуемость процесса, исчезают «подвешенные» решения и затяжные согласования.

Выдержка из книги: [docs/source/book.md#L2416-L2830](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2416-L2830); [docs/source/book.md#L2924-L2968](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L2924-L2968); [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918)

Фрагмент документа для изменения: [docs/design-guidelines/visual/AGENTS.md#L1-L15](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/visual/AGENTS.md#L1-L15) (ориентир: `Документ целиком`).

### 3. Сделать требования к качеству документа проверяемыми
Что доработать: Перевести ключевые пункты из декларативной формы в проверяемые критерии: что именно считается выполнением требования.

Как внедрить: Для каждого правила добавить поле «проверка/артефакт доказательства» (например: ссылка на тест, линтер, checklist, отчёт).

Ожидаемый эффект: Чек-листы становятся объективными, а качество исполнения можно измерять и сравнивать между PR.

Выдержка из книги: [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574); [docs/source/book.md#L13575-L14319](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13575-L14319); [docs/source/book.md#L21983-L22604](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21983-L22604)

Фрагмент документа для изменения: [docs/design-guidelines/visual/AGENTS.md#L1-L15](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/visual/AGENTS.md#L1-L15) (ориентир: `Документ целиком`).

### 4. Встроить трассируемость от стандарта к артефактам разработки
Что доработать: Добавить явные ссылки: правило -> шаблон документа -> код/контракт -> тест/метрика.

Как внедрить: Ввести идентификаторы правил (например DG-001, DG-002) и требовать их указывать в PR/Issue/ADR при изменениях.

Ожидаемый эффект: Упрощается анализ влияния изменений и поддержка соответствия стандартам при эволюции платформы.

Выдержка из книги: [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311); [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062)

Фрагмент документа для изменения: [docs/design-guidelines/visual/AGENTS.md#L1-L15](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/visual/AGENTS.md#L1-L15) (ориентир: `Документ целиком`).

### 5. Описать управляемый процесс изменения инструкций
Что доработать: Добавить политику изменения гайда: как подаётся запрос, какие атрибуты обязательны, как проводится анализ влияния и как документируется решение.

Как внедрить: Вынести минимальный workflow: draft -> review -> approved/rejected -> effective date, плюс журнал изменений.

Ожидаемый эффект: Изменения стандартов перестанут быть ad-hoc, уменьшится риск регрессий в процессах и документации.

Выдержка из книги: [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109)

Фрагмент документа для изменения: [docs/design-guidelines/visual/AGENTS.md#L1-L15](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/visual/AGENTS.md#L1-L15) (ориентир: `Документ целиком`).

### 6. Добавить риск-модель для нарушений стандартов
Что доработать: Зафиксировать типовые риски: пропуск обязательных гайдов, несогласованные исключения, потеря совместимости документов.

Как внедрить: Для каждого риска указать причину, эффект, ранний индикатор и меру предотвращения.

Ожидаемый эффект: Команда сможет заранее видеть уязвимые места процесса и реагировать до появления системных дефектов.

Выдержка из книги: [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510); [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610)

Фрагмент документа для изменения: [docs/design-guidelines/visual/AGENTS.md#L1-L15](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/visual/AGENTS.md#L1-L15) (ориентир: `Документ целиком`).
