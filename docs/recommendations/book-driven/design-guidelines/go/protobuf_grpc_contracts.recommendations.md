# Рекомендации по доработке `docs/design-guidelines/go/protobuf_grpc_contracts.md`

- Дорабатываемый документ: [docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29)
- Профиль рекомендаций: `guideline_go`
- Дата генерации: `2026-02-06`

## Рекомендации

### 1. Фиксировать источник и обоснование для каждого обязательного правила
Что доработать: Для правил по transport/repository/contracts добавить поле «почему это нужно» и «какой дефект предотвращает».

Как внедрить: Использовать единый шаблон: правило -> источник требования -> риск при нарушении -> ссылка на проверку.

Ожидаемый эффект: Это повышает дисциплину исполнения и помогает быстрее оценивать необходимость исключений.

Выдержка из книги: [docs/source/book.md#L13373-L13574](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13373-L13574); [docs/source/book.md#L29827-L29918](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29827-L29918); [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355)

Фрагмент документа для изменения: [docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29) (ориентир: `Protobuf/gRPC: правила транспортных контрактов`).

### 2. Сделать нормы по контрактам и совместимости проверяемыми
Что доработать: Добавить для каждого протокола явные критерии совместимости и сценарий обратной совместимости при изменениях.

Как внедрить: Включить обязательную связку: изменение контракта -> изменение тестов -> изменение миграции/документации.

Ожидаемый эффект: Снижается вероятность регрессий на интеграциях и «тихих» breaking changes.

Выдержка из книги: [docs/source/book.md#L12254-L12670](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L12254-L12670); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062); [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594)

Фрагмент документа для изменения: [docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29) (ориентир: `Protobuf/gRPC: правила транспортных контрактов`).

### 3. Добавить контрольные списки дефектов для типовых Go-ошибок в требованиях
Что доработать: Внести в каждый профильный документ раздел «анти-паттерны/дефекты», которые нужно проверять на ревью.

Как внедрить: Сопроводить раздел примерами «до/после» для неоднозначных требований и описаний поведения.

Ожидаемый эффект: Ревью становится системным, а качество проектирования сервисов выше при меньшем числе итераций.

Выдержка из книги: [docs/source/book.md#L21983-L22604](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21983-L22604); [docs/source/book.md#L13575-L14319](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L13575-L14319); [docs/source/book.md#L23123-L24148](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L23123-L24148)

Фрагмент документа для изменения: [docs/design-guidelines/go/protobuf_grpc_contracts.md#L23-L29](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/go/protobuf_grpc_contracts.md#L23-L29) (ориентир: `Ошибки в gRPC`).

### 4. Связать технические правила с этапами жизненного цикла требований
Что доработать: Отдельно отметить, на каком этапе используется правило: анализ, дизайн, реализация, тестирование, эксплуатация.

Как внедрить: Для этапов добавить обязательные входные и выходные критерии.

Ожидаемый эффект: Документы станут пригодны для планирования работ, а не только как справочник по стилю.

Выдержка из книги: [docs/source/book.md#L24411-L24599](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24411-L24599); [docs/source/book.md#L24600-L24943](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24600-L24943); [docs/source/book.md#L24944-L25062](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L24944-L25062)

Фрагмент документа для изменения: [docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29) (ориентир: `Документ целиком`).

### 5. Ввести политику управляемых изменений правил Go-гайдов
Что доработать: Определить процесс запроса на изменение правила, обязательный impact-анализ и критерии отката.

Как внедрить: Добавить атрибуты версии и статуса у каждого правила, затрагивающего контракты и инфраструктуру.

Ожидаемый эффект: Изменения стандартов станут прогнозируемыми, а внедрение новых правил будет контролируемым.

Выдержка из книги: [docs/source/book.md#L30646-L31033](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30646-L31033); [docs/source/book.md#L30558-L30594](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L30558-L30594); [docs/source/book.md#L29919-L30109](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L29919-L30109)

Фрагмент документа для изменения: [docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/go/protobuf_grpc_contracts.md#L1-L29) (ориентир: `Protobuf/gRPC: правила транспортных контрактов`).

### 6. Добавить риск-каталог для критичных инженерных зон
Что доработать: Собрать риски по темам: миграции, idempotency, error handling, observability, security; для каждого риска прописать детектор и mitigation.

Как внедрить: Ссылаться на риск-каталог из профильных правил, чтобы ревью покрывало не только стиль, но и эксплуатационную безопасность.

Ожидаемый эффект: Это сократит число production-инцидентов, вызванных неполными инженерными требованиями.

Выдержка из книги: [docs/source/book.md#L35136-L35355](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L35136-L35355); [docs/source/book.md#L18065-L19261](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L18065-L19261); [docs/source/book.md#L21321-L21798](https://github.com/codex-k8s/codex-k8s/blob/main/docs/source/book.md#L21321-L21798)

Фрагмент документа для изменения: [docs/design-guidelines/go/protobuf_grpc_contracts.md#L23-L29](https://github.com/codex-k8s/codex-k8s/blob/main/docs/design-guidelines/go/protobuf_grpc_contracts.md#L23-L29) (ориентир: `Ошибки в gRPC`).
