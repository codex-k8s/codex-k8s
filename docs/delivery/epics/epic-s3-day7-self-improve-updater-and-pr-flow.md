---
doc_id: EPC-CK8S-S3-D7
type: epic
title: "Epic S3 Day 7: run:self-improve updater and PR flow"
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

# Epic S3 Day 7: run:self-improve updater and PR flow

## TL;DR
- Цель: превратить self-improve diagnostics в управляемые изменения.
- MVP-результат: агент генерирует PR с улучшениями инструкций/документации/шаблонов, прикладывает rationale и evidence.

## Priority
- `P0`.

## Scope
### In scope
- Автогенерация изменений в docs/prompt seeds/agent instructions по approved action-plan.
- PR flow с traceability: что исправлено, из каких run/log/comment выводов.
- Guardrails против деградации стандартов (checks против ослабления policy/security).
- Привязка результата к исходному issue/pr через `links`.

### Out of scope
- Автоматический merge без review.

## Критерии приемки
- Минимум один self-improve PR создаётся end-to-end с проверяемым улучшением и понятной аргументацией.
