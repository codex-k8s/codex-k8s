---
doc_id: EPC-CK8S-S2-D2
type: epic
title: "Epic S2 Day 2: Issue label triggers for run:dev and run:dev:revise"
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

# Epic S2 Day 2: Issue label triggers for run:dev and run:dev:revise

## TL;DR
- Цель эпика: сделать GitHub Issue лейблы входом для запуска разработки внутри `codex-k8s`.
- Ключевая ценность: разработка становится webhook-driven и трассируемой (flow_events + run state).
- MVP-результат: `issues.labeled` webhook создаёт run request и ставит run в очередь на исполнение.

## Priority
- `P0`.

## Scope
### In scope
- Поддержка GitHub webhook события `issues` (label added).
- Правила авторизации для trigger-лейблов (`run:*`):
  - учитываем политику “trigger labels только через апрув Owner” (как принцип);
  - на MVP: allowlist/роль, проверка sender в webhook payload, запись audit события.
- Зафиксировать полный каталог `run:*`, `state:*`, `need:*` в документации и GitHub vars (даже если часть run labels пока не активна).
- Маппинг лейблов:
  - `run:dev` -> создать dev run;
  - `run:dev:revise` -> запустить revise run (на существующий PR/ветку).
- Запись событий в `flow_events` и создание/обновление записи run/queue.

### Out of scope
- Автоматическое назначение/снятие лейблов агентом без политики/апрувов.

## Data model impact
- Добавление таблицы/полей для “run request” (если текущая `agent_runs` модель не покрывает issue-driven use-case).
- Индексы: по `(repo, issue_number, kind, status)` или эквивалент.

## Критерии приемки эпика
- Добавление лейбла `run:dev` на Issue приводит к созданию run request и появлению в UI/логах.
- Несанкционированный actor не может триггерить запуск (событие отклоняется и логируется).
- Workflow-условия для активных labels используют `vars.*`, а не строковые литералы.
