---
doc_id: EPC-CK8S-S3-D1
type: epic
title: "Epic S3 Day 1: Full stage and label activation"
status: planned
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S3 Day 1: Full stage and label activation

## TL;DR
- Цель: включить весь каталог `run:*`/`state:*`/`need:*` как реально исполняемую stage-модель, а не только документный план.
- MVP-результат: каждый stage имеет валидный trigger path, policy и traceability.

## Priority
- `P0`.

## Scope
### In scope
- Активация trigger-label обработки для `run:intake..run:ops`, `run:*:revise`, `run:abort`, `run:rethink`, `run:self-improve`.
- Обновление state machine переходов между стадиями и revise/rollback петлями.
- Валидация конфликтов labels и preconditions на вход stage.
- Синхронизация labels-as-vars каталога и audit событий.
- При ошибках валидации labels — локализованная отбивка под Issue/PR по шаблону run-status:
  - какие labels конфликтуют;
  - просьба снять конфликтующие labels и оставить один валидный trigger.

### Out of scope
- Глубокая бизнес-логика каждого stage (дорабатывается по следующим эпикам).

## Критерии приемки
- Все stage labels маршрутизируются и пишут события переходов в audit.
- Ошибочные/конфликтные переходы отклоняются детерминированно с диагностикой.
- Для конфликтов labels публикуется человекочитаемое сообщение в Issue/PR с конкретным remediation-шагом.
