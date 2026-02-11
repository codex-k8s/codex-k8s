---
doc_id: EPC-CK8S-S2-D2
type: epic
title: "Epic S2 Day 2: Issue label triggers for run:dev and run:dev:revise"
status: completed
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-12
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-11-s2-day2-complete"
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
- Предусловие: OpenAPI contract-first baseline из S2 Day1 внедрён и используется как source of truth для HTTP transport.
- Поддержка GitHub webhook события `issues` (label added).
- Правила авторизации для trigger-лейблов (`run:*`):
  - учитываем политику “trigger labels только через апрув Owner” (как принцип);
  - на MVP: allowlist/роль, проверка sender в webhook payload, запись audit события.
- Зафиксировать полный каталог labels в документации и GitHub vars:
  - `run:*`, `state:*`, `need:*`;
  - конфигурационные `[ai-model-*]`, `[ai-reasoning-*]`.
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

## Прогресс реализации (2026-02-11)
- Реализован ingest `issues.labeled` в `control-plane`:
  - trigger labels `run:dev` и `run:dev:revise` читаются из env (`CODEXK8S_RUN_DEV_LABEL`, `CODEXK8S_RUN_DEV_REVISE_LABEL`);
  - нетриггерные issue label события помечаются как `ignored` с записью `webhook.ignored` в `flow_events`.
- Добавлена авторизация sender для issue-trigger:
  - разрешены `platform_owner`, `platform_admin`, `project_member` c ролями `admin|read_write`;
  - неразрешённые попытки фиксируются в `flow_events` с причиной.
- Расширен payload run/event:
  - в `agent_runs.run_payload` добавляются `issue` + `trigger` metadata;
  - в `flow_events.payload` добавляются `action/sender/repository/label/run_kind`.
- Обновлён HTTP контракт webhook ingest:
  - `status` enum: `accepted|duplicate|ignored`;
  - `api-gateway` возвращает `200` для `duplicate|ignored`, `202` для `accepted`.
- Bootstrap/deploy синхронизация:
  - `CODEXK8S_GITHUB_WEBHOOK_EVENTS` включает `issues`;
  - каталог vars `run:*|state:*|need:*` синхронизируется в GitHub Variables;
  - runtime secret пополнен `CODEXK8S_RUN_DEV_LABEL` и `CODEXK8S_RUN_DEV_REVISE_LABEL`.
