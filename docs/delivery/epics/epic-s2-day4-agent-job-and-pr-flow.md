---
doc_id: EPC-CK8S-S2-D4
type: epic
title: "Epic S2 Day 4: Agent job image, git workflow and PR creation"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-10
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 4: Agent job image, git workflow and PR creation

## TL;DR
- Цель эпика: довести run до результата “создан PR” (для dev) и “обновлён PR” (для revise).
- Ключевая ценность: полный dogfooding цикл без ручного вмешательства.
- MVP-результат: агентный Job клонирует repo, вносит изменения, пушит ветку и открывает PR.

## Priority
- `P0`.

## Scope
### In scope
- Определить image/entrypoint агентного Job (инструменты: git, gh, go/node при необходимости).
- Политика кредов:
  - repo token берётся из БД (шифрованно), расшифровывается в control-plane и прокидывается в Job безопасно;
  - исключить попадание токенов в логи.
- PR flow:
  - детерминированное имя ветки;
  - создание PR с ссылкой на Issue;
  - запись PR URL/номер в БД.

### Out of scope
- Автоматический code review (финальный ревью остаётся за Owner).

## Критерии приемки эпика
- `run:dev` создаёт PR.
- `run:dev:revise` обновляет существующий PR (или создаёт новый по политике).
- В `flow_events` есть трасса: issue -> run -> namespace -> job -> pr.

