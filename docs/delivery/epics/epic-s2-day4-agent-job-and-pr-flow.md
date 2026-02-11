---
doc_id: EPC-CK8S-S2-D4
type: epic
title: "Epic S2 Day 4: Agent job image, git workflow and PR creation"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-11
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
- Policy шаблонов промптов:
  - `work`/`review` шаблоны для запуска берутся по приоритету `DB override -> repo seed`;
  - шаблоны выбираются по locale policy `project locale -> system default -> en`;
  - для системных агентов baseline заполняется минимум `ru` и `en`;
  - для run фиксируется effective template source/version/locale в аудит-контуре.
- Resume policy:
  - сохранять `codex-cli` session JSON в `agent_sessions`;
  - при перезапуске/возобновлении run восстанавливать сессию с того же места.
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
