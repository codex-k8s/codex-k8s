---
doc_id: EPC-CK8S-S3-D1
type: epic
title: "Epic S3 Day 1: Full stage and label activation"
status: completed
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
- Активация trigger-label обработки для `run:intake..run:ops`, `run:*:revise`, `run:rethink`, `run:self-improve`.
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

## Фактический результат (выполнено)
- Активирован полный каталог trigger-kind и labels:
  - `run:intake..run:ops`,
  - revise-контур `run:<stage>:revise`,
  - служебные `run:rethink`, `run:self-improve`.
- Включена унифицированная нормализация/проверка trigger-kind в shared domain:
  - `NormalizeTriggerKind`,
  - `IsKnownTriggerKind`,
  - `IsReviseTriggerKind`,
  - `DefaultTriggerLabel`.
- В webhook ingestion реализована детерминированная обработка конфликтных `run:*` labels:
  - run не создаётся;
  - фиксируется `webhook.ignored` с reason `issue_trigger_label_conflict`;
  - публикуется локализованная диагностическая отбивка в Issue через шаблон run-status.
- Введено новое audit-событие `run.trigger.conflict.comment` для трассировки публикации конфликтной диагностики.
- Runtime-профиль и template selection синхронизированы с full-stage моделью:
  - `full-env` применяется для всех известных stage-trigger;
  - revise-шаблон выбирается по общему правилу revise-trigger (не только `dev_revise`).
- Каталог labels-as-vars синхронизирован в bootstrap/deploy/workflows:
  - добавлен `CODEXK8S_RUN_SELF_IMPROVE_LABEL`,
  - полный каталог `CODEXK8S_RUN_* / CODEXK8S_STATE_* / CODEXK8S_NEED_*` прокинут в deploy pipeline и ConfigMap.
- Для control-plane/worker/agent-runner добавлены/обновлены тесты на full-stage trigger routing и conflict handling.

## Проверки
- `make lint-go` — passed.
- `make dupl-go` — passed.
- `go test ./libs/... ./services/internal/control-plane/... ./services/jobs/worker/... ./services/jobs/agent-runner/...` — passed.
