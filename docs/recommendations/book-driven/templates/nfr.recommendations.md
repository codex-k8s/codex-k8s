# Рекомендации по доработке `docs/templates/nfr.md`

- Дорабатываемый документ: [docs/templates/nfr.md#L1-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/nfr.md#L1-L60)
- Профиль рекомендаций: `template_architecture`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Добавить обязательный источник требований для архитектурных решений
Что доработать: В шаблон включить секцию «источник требования»: ID из PRD/NFR/регуляторики/инцидентов с явным обоснованием.

Как внедрить: Запретить архитектурные решения без явной привязки к исходному требованию и границе системы.

Ожидаемый эффект: Это делает архитектурный документ проверяемым и снижает риск «дизайна ради дизайна».

Выдержка из книги: [docs/source/book.md#L12254-L12670](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L12254-L12670); [docs/source/book.md#L5401-L6092](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L5401-L6092); [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918)

Фрагмент документа для изменения: [docs/templates/nfr.md#L25-L27](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/nfr.md#L25-L27) (ориентир: `Контекст`).

### 2. Стандартизовать сравнение альтернатив через ценность/стоимость/риск
Что доработать: Для каждого варианта добавить явные оценки value, implementation cost, operational risk и time-to-market.

Как внедрить: Фиксировать критерий выбора и причину отклонения альтернатив в сопоставимом формате.

Ожидаемый эффект: Повышается прозрачность архитектурных компромиссов и управляемость техдолга.

Выдержка из книги: [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798); [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L6532-L6610](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L6532-L6610)

Фрагмент документа для изменения: [docs/templates/nfr.md#L1-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/nfr.md#L1-L60) (ориентир: `Документ целиком`).

### 3. Добавить измеримые NFR и критерии приемки архитектуры
Что доработать: Каждое качество (надёжность, производительность, безопасность, observability) оформить как измеримое требование с порогами.

Как внедрить: Указать источник метрики, способ проверки и шаги при отклонении от порога.

Ожидаемый эффект: Документ становится практичным инструментом для валидации архитектурной реализуемости.

Выдержка из книги: [docs/source/book.md#L18065-L19261](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L18065-L19261); [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574); [docs/source/book.md#L22946-L23122](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L22946-L23122)

Фрагмент документа для изменения: [docs/templates/nfr.md#L17-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/nfr.md#L17-L60) (ориентир: `NFR: <Категория> — <Система/Фича>`).

### 4. Встроить матрицу трассируемости требований к компонентам, коду и тестам
Что доработать: Добавить таблицу связей: requirement ID -> компонент/интерфейс -> тест/мониторинг.

Как внедрить: Требовать обновления матрицы при каждом изменении контракта, миграции или поведения.

Ожидаемый эффект: Снижается вероятность пропуска критичных зависимостей при доработках.

Выдержка из книги: [docs/source/book.md#L32044-L32311](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L32044-L32311); [docs/source/book.md#L24600-L24943](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24600-L24943); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062)

Фрагмент документа для изменения: [docs/templates/nfr.md#L45-L48](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/nfr.md#L45-L48) (ориентир: `Тестирование и верификация`).

### 5. Добавить lifecycle-атрибуты решений и требований
Что доработать: В шаблон включить атрибуты: версия, статус, владелец, дата пересмотра, целевой релиз.

Как внедрить: Использовать атрибуты в отчётах готовности и при планировании релизов.

Ожидаемый эффект: Упрощается управление большим портфелем архитектурных решений.

Выдержка из книги: [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109); [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599)

Фрагмент документа для изменения: [docs/templates/nfr.md#L1-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/nfr.md#L1-L60) (ориентир: `Документ целиком`).

### 6. Описать процесс изменения и отката архитектурных решений
Что доработать: Добавить формальный workflow изменений: запрос -> impact analysis -> решение -> реализация -> проверка -> откат при необходимости.

Как внедрить: Фиксировать критерии запуска отката и ответственность за решение.

Ожидаемый эффект: Снижается риск неуправляемых изменений в критичных архитектурных зонах.

Выдержка из книги: [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033); [docs/source/book.md#L21983-L22604](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21983-L22604)

Фрагмент документа для изменения: [docs/templates/nfr.md#L57-L60](https://github.com/codex-k8s/codex-k8s/blob/main/docs/templates/nfr.md#L57-L60) (ориентир: `Апрув`).
