---
doc_id: ARC-C4C-XXXX
type: c4-context
title: "<Система> — C4 Context"
status: draft
owner_role: SA
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
related_issues: []
related_prs: []
approvals:
  required: ["CTO"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# C4 Context: <Система>

## TL;DR
- Система в контуре: ...
- Пользователи: ...
- Внешние зависимости: ...

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title <Система> - System Context

Person(user, "User", "Кто это и зачем")
System(system, "<Система>", "Краткое описание")

System_Ext(ext1, "External System 1", "Что это")
System_Ext(ext2, "External System 2", "Что это")

Rel(user, system, "Uses", "HTTP/APP")
Rel(system, ext1, "Calls", "HTTPS")
Rel(system, ext2, "Publishes/Consumes", "Kafka")
```

## Пояснения

- Основные взаимодействия:
- Границы ответственности:

## Внешние зависимости

- ext1: SLA/SLO? Риски?
- ext2: ...

## Открытые вопросы

- ...

## Апрув

- request_id: ...
- Решение:
- Комментарий:
