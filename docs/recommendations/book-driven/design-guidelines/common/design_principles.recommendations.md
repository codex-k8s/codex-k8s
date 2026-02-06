# Рекомендации по доработке `docs/design-guidelines/common/design_principles.md`

- Дорабатываемый документ: [docs/design-guidelines/common/design_principles.md#L1-L58](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/common/design_principles.md#L1-L58)
- Профиль рекомендаций: `guideline_common`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Явно связать архитектурные принципы с бизнес-целями и границами системы
Что доработать: В каждом принципе указать, какую бизнес-цель он защищает и какие изменения считаются выходом за рамки.

Как внедрить: Добавить мини-матрицу: принцип -> ожидаемый эффект -> ограничение -> допустимое исключение.

Ожидаемый эффект: Документ станет не набором лозунгов, а рабочим инструментом принятия архитектурных решений.

Выдержка из книги: [docs/source/book.md#L5401-L6092](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L5401-L6092); [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610); [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599)

Фрагмент документа для изменения: [docs/design-guidelines/common/design_principles.md#L1-L58](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/common/design_principles.md#L1-L58) (ориентир: `Принципы проектирования`).

### 2. Добавить критерии качества формулировок правил
Что доработать: Для неоднозначных формулировок ввести «словарь трактовок» и конкретизировать проверяемые условия соблюдения.

Как внедрить: Отдельно отметить запрещённые двусмысленные формулировки и заменить их на измеримые критерии.

Ожидаемый эффект: Это снизит риск разночтений между командами и ускорит ревью архитектурных изменений.

Выдержка из книги: [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574); [docs/source/book.md#L13575-L14319](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13575-L14319); [docs/source/book.md#L21983-L22604](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21983-L22604)

Фрагмент документа для изменения: [docs/design-guidelines/common/design_principles.md#L1-L58](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/common/design_principles.md#L1-L58) (ориентир: `Принципы проектирования`).

### 3. Усилить секцию об использовании общих библиотек через правила повторного применения
Что доработать: Добавить критерии, когда правило/компонент переносится в `libs/*`, а когда остаётся локальным.

Как внедрить: Включить требования к обратной совместимости и миграционному плану при изменении общих библиотек.

Ожидаемый эффект: Снижается техдолг от дублирования и риск хаотичного роста shared-кода.

Выдержка из книги: [docs/source/book.md#L23123-L24148](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L23123-L24148); [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033)

Фрагмент документа для изменения: [docs/design-guidelines/common/design_principles.md#L31-L36](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/common/design_principles.md#L31-L36) (ориентир: `DRY`).

### 4. Добавить трассируемость к контрактам и тестам
Что доработать: Для архитектурных ограничений определить обязательные контрольные точки: контракт, тест, метрика или линтер.

Как внедрить: Зафиксировать минимальный набор доказательств соблюдения на уровне PR.

Ожидаемый эффект: Упростится контроль исполнения принципов и аудит архитектурного качества.

Выдержка из книги: [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062); [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510)

Фрагмент документа для изменения: [docs/design-guidelines/common/design_principles.md#L1-L58](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/common/design_principles.md#L1-L58) (ориентир: `Документ целиком`).

### 5. Добавить атрибуты состояния требований к архитектуре
Что доработать: Ввести поля статуса для архитектурных ограничений: proposed/active/deprecated/superseded и дату ревизии.

Как внедрить: Связать эти статусы с ритмом пересмотра (например, раз в квартал) и журналом изменений.

Ожидаемый эффект: Это предотвращает «забытые» правила и устаревшие ограничения в базе знаний команды.

Выдержка из книги: [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109); [docs/source/book.md#L34311-L34510](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L34311-L34510)

Фрагмент документа для изменения: [docs/design-guidelines/common/design_principles.md#L1-L58](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/common/design_principles.md#L1-L58) (ориентир: `Документ целиком`).

### 6. Встроить риск-подход в архитектурные руководства
Что доработать: Для каждого критичного ограничения указать типовые риски нарушения и минимальную стратегию смягчения.

Как внедрить: Добавить привязку рисков к чек-листам и шаблонам документов, где эти риски должны контролироваться.

Ожидаемый эффект: Команда будет заранее видеть потенциальные сбои в архитектуре и реагировать до инцидентов.

Выдержка из книги: [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033)

Фрагмент документа для изменения: [docs/design-guidelines/common/design_principles.md#L1-L58](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/common/design_principles.md#L1-L58) (ориентир: `Документ целиком`).
