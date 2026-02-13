---
doc_id: EPC-CK8S-S3-D6
type: epic
title: "Epic S3 Day 6: run:self-improve ingestion and diagnostics"
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
