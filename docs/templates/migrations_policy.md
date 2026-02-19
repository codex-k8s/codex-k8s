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
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# DB Migrations Policy: <Система>

## TL;DR
- Подход: backward-compatible / expand-contract / zero-downtime
- Инструменты миграций:
- Где лежат миграции и кто владелец схемы:
- Политика откатов:

## Размещение миграций и владелец схемы

Правила:
- В монорепо миграции живут *внутри держателя схемы*, например:
  `services/<zone>/<db-owner-service>/cmd/cli/migrations/*.sql`.
- Shared DB без владельца запрещён: если БД/схема общая для нескольких сервисов, всё равно должен быть один владелец миграций и схемы.
- История миграций не переписывается; новые изменения только новыми файлами.

## Принципы
- Нулевой даунтайм (если требуется):
- Backward compatibility:
- Версионирование:

## Процесс миграции (шаги)
1) Expand (добавляем поля/таблицы)
2) Dual-write / Backfill (если нужно)
3) Switch reads
4) Contract (удаляем старое после проверки)

## Как выполняются миграции при деплое
- Production/Prod стратегия:
- Гарантия отсутствия параллельных миграций (например, отдельный Kubernetes Job или initContainer + advisory lock):
- Поведение при ошибке миграции:

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
