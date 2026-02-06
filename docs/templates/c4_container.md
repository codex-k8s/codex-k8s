---
doc_id: ARC-C4N-XXXX
type: c4-container
title: "<Система> — C4 Container"
status: draft
owner_role: SA
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
related_issues: []
related_prs: []
approvals:
  required: ["CTO"]
  status: pending
  request_id: ""
---

# C4 Container: <Система>

## TL;DR
- Основные контейнеры: ...
- Технологии: ...
- Потоки данных: ...

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title <Система> - Container Diagram

Person(user, "User", "...")
System_Boundary(b0, "<Система>") {
  Container(web, "Web/UI", "React", "Frontend")
  Container(api, "API", "Go", "REST/gRPC")
  ContainerDb(db, "DB", "PostgreSQL", "Данные")
  Container(queue, "Queue", "Kafka", "События")
}

System_Ext(ext1, "External System", "...")

Rel(user, web, "Uses", "HTTPS")
Rel(web, api, "Calls", "HTTPS")
Rel(api, db, "Reads/Writes", "SQL")
Rel(api, queue, "Publishes", "Kafka")
Rel(api, ext1, "Calls", "HTTPS")
```

## Контейнеры (описание)

### Web/UI

* Ответственность:
* Деплой:
* Риски:

### API

* Ответственность:
* Контракты:
* Ограничения:

### DB

* Схема/миграции:
* Резервирование/бэкап:

### Queue

* Топики:
* Гарантии доставки:

## Открытые вопросы

* ...

## Апрув

* request_id: ...
* Решение:
* Комментарий:
