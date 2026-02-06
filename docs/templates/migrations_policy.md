---
doc_id: MIG-XXXX
type: migrations-policy
title: "<Система> — DB Migrations Policy"
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

# DB Migrations Policy: <Система>

## TL;DR
- Подход: backward-compatible / expand-contract / zero-downtime
- Инструменты миграций:
- Политика откатов:

## Принципы
- Нулевой даунтайм (если требуется):
- Backward compatibility:
- Версионирование:

## Процесс миграции (шаги)
1) Expand (добавляем поля/таблицы)
2) Dual-write / Backfill (если нужно)
3) Switch reads
4) Contract (удаляем старое после проверки)

## Политика backfill
- Как выполняем:
- Ограничение по скорости:
- Мониторинг прогресса:

## Политика rollback
- Когда можно rollback:
- Что нельзя откатить:
- План отката (ссылка на rollback_plan.md):

## Проверки
- Pre-migration checks:
- Post-migration verification:

## Открытые вопросы
- ...

## Апрув
- request_id: ...
- Решение:
- Комментарий:
