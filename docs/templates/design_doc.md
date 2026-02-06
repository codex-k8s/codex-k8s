---
doc_id: DSG-XXXX
type: design-doc
title: "<Название> — Detailed Design"
status: draft
owner_role: SA
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
related_issues: []
related_prs: []
related_adrs: []
approvals:
  required: ["CTO"]
  status: pending
  request_id: ""
---

# Detailed Design: <Название>

## TL;DR
- Что меняем:
- Почему:
- Основные компоненты:
- Риски:
- План выката:

## Цели / Не-цели
### Goals
- ...

### Non-goals
- ...

## Контекст и текущая архитектура
- Ссылки на C4:
- Болевые точки:

## Предлагаемый дизайн (high-level)
- Диаграмма (ссылка/mermaid):
- Компоненты:
- Потоки данных:

## API/Контракты
- Ссылка на OpenAPI/proto:
- Изменения совместимости:
- Ошибки/ретраи:

## Модель данных и миграции
- Сущности:
- Миграции:
- Backfill:
- Rollback constraints:

## Сценарии (Sequence diagrams)
```mermaid
sequenceDiagram
  participant U as User
  participant A as API
  participant D as DB
  U->>A: Request
  A->>D: Query/Write
  D-->>A: Result
  A-->>U: Response
````

## Нефункциональные аспекты

* Надёжность:
* Производительность:
* Безопасность:
* Наблюдаемость:

## Наблюдаемость (Observability)

* Логи:
* Метрики:
* Трейсы:
* Дашборды:
* Алерты:

## Тестирование

* Юнит:
* Интеграция:
* E2E:
* Нагрузочное:
* Security checks:

## План выката (Rollout)

* Staging:
* Canary/gradual rollout:
* Feature flags:
* План коммуникаций:

## План отката (Rollback)

* Триггеры:
* Шаги:
* Проверка успеха:

## Альтернативы и почему отвергли

* ...

## Открытые вопросы (для Telegram Executor)

1. ...

## Апрув

* request_id: ...
* Решение:
* Комментарий:
