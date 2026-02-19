---
doc_id: EPC-CK8S-S3-D7
type: epic
title: "Epic S3 Day 7: run:self-improve updater and PR flow"
status: completed
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
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
- Подготовка и сопровождение минимальной stage-matrix prompt seeds (`docs/product/prompt-seeds/<stage>-work.md`, `<stage>-review.md`) для изоляции dev-шаблона при тестировании остальных stage-run.

### Planned follow-up (post-MVP hardening)
- Комплексная проработка role-specific prompt matrix:
  - отдельные `work/review` шаблоны для всех системных ролей (`pm`, `sa`, `em`, `dev`, `reviewer`, `qa`, `sre`, `km`);
  - отдельные шаблоны для специальных режимов (`mode:discussion`, `run:self-improve`);
  - унификация locale-пакетов (`ru`/`en`) и проверка консистентности policy-blocks;
  - автоматические quality checks для prompt templates (lint/validation/traceability coverage).

### Out of scope
- Автоматический merge без review.

## Критерии приемки
- Минимум один self-improve PR создаётся end-to-end с проверяемым улучшением и понятной аргументацией.

## Фактический результат (выполнено)
- Для `run:self-improve` закреплён review-контур шаблонов:
  - в runtime self-improve trigger всегда использует `template_kind=review`.
- В stage prompt matrix добавлен self-improve review seed:
  - `docs/product/prompt-seeds/self-improve-review.md`.
- Обновлены seed-инструкции self-improve work/review:
  - сначала `AGENTS.md`, затем Issue/comments, затем связанная документация;
  - обязательная MCP-диагностика run/session (`self_improve_runs_list`, `self_improve_run_lookup`, `self_improve_session_get`);
  - обязательное сохранение session JSON во временный каталог `/tmp/codex-sessions/<run-id>`;
  - обязательный traceability output (`diagnosis`, `action_items`, `evidence_refs`, `tool_gaps`);
  - требования к PR flow, проверкам и ограничениям policy/security.
- В runner-контракт добавлены поля structured output для self-improve updater:
  - `diagnosis`, `action_items`, `evidence_refs`, `tool_gaps`.
- Session snapshot сохраняет structured self-improve поля в `session_json`, что упрощает последующий audit/review цикл.

## Проверки
- `go test ./services/jobs/worker/internal/domain/worker` — passed.
- `go test ./services/jobs/agent-runner/internal/runner` — passed.
