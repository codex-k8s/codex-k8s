---
doc_id: EPC-CK8S-S3-D6
type: epic
title: "Epic S3 Day 6: run:self-improve ingestion and diagnostics"
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

# Epic S3 Day 6: run:self-improve ingestion and diagnostics

## TL;DR
- Цель: запустить ingest-контур самоулучшения по лейблу `run:self-improve`.
- MVP-результат: агент собирает и нормализует run-логи, комментарии Owner/бота, PR артефакты и формирует improvement diagnosis.

## Priority
- `P0`.

## Scope
### In scope
- Trigger path для `run:self-improve` и policy preconditions.
- Сбор входных данных: `agent_sessions`, `flow_events`, PR/Issue comments, связанный diff и артефакты.
- Диагностика повторяющихся проблем и формирование action-plan.
- Классификация рекомендаций: docs, prompts, instructions, tooling/image.

### Out of scope
- Автоматическое применение всех предложений без review/approval.

## Критерии приемки
- После запуска `run:self-improve` формируется структурированный отчёт с actionable items и трассировкой источников.

## Фактический результат (выполнено)
- Маршрутизация `run:self-improve` закреплена за системной ролью `km`:
  - webhook trigger `self_improve` теперь резолвится в `agent_key=km`.
- Добавлена миграция системного role-catalog для stage-flow:
  - `pm`, `sa`, `em`, `sre`, `km` (в дополнение к уже существующим `dev`, `reviewer`, `qa`).
- Для self-improve запуска включена обязательная структурированная диагностика:
  - `diagnosis`;
  - `action_items[]`;
  - `evidence_refs[]`;
  - `tool_gaps[]` (опционально).
- В MCP добавлены read-only диагностические ручки self-improve:
  - `self_improve_runs_list` (history pagination 50/page, newest-first);
  - `self_improve_run_lookup` (поиск run по Issue/PR);
  - `self_improve_session_get` (получение `codex-cli` session JSON и target path в `/tmp/codex-sessions/...`).
- При self-improve run публикуется audit-событие:
  - `run.self_improve.diagnosis_ready`.
- Prompt-body для self-improve актуализирован под ingestion-диагностику:
  - обязательный сбор входов из `agent_sessions`, `flow_events`, Issue/PR comments и артефактов;
  - классификация рекомендаций по категориям (`docs`, `prompts`, `instructions`, `tooling/image`).

## Проверки
- `go test ./services/internal/control-plane/internal/domain/webhook` — passed.
- `go test ./services/jobs/worker/internal/domain/worker` — passed.
- `go test ./services/jobs/agent-runner/internal/runner` — passed.
