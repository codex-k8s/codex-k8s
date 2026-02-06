---
doc_id: RLB-XXXX
type: rollback-plan
title: "<Система/Версия> — Rollback Plan"
status: draft
owner_role: SRE
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
related_prs: []
approvals:
  required: ["CTO"]
  status: pending
  request_id: ""
---

# Rollback Plan: <Система> <Версия>

## TL;DR
- Когда откатываем:
- Как откатываем:
- Как проверяем:

## Триггеры rollback
- SLO burn rate превышен:
- Рост ошибок:
- Деградация latency:
- Сигналы от бизнеса:

## Предусловия
- Бэкапы/снапшоты:
- Совместимость схемы:
- Наличие предыдущего стабильного релиза:

## Варианты отката
### 1) Откат приложения (без схемы)
- Шаги:
- Риски:

### 2) Откат приложения + схемы
- Шаги:
- Ограничения/риски:

### 3) Kill switch / feature flag
- Шаги:

## Пошаговая инструкция rollback (runbook-style)
1) ...
2) ...
3) ...

## Верификация после rollback
- Проверки функционала:
- Метрики:
- Логи:
- Smoke тесты:

## Коммуникации
- Кто уведомляется:
- Шаблон сообщения:

## Апрув (обязателен)
- request_id: ...
- Решение:
- Комментарий:
