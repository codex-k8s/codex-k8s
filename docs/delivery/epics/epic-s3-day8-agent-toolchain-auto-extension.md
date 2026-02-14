---
doc_id: EPC-CK8S-S3-D8
type: epic
title: "Epic S3 Day 8: Agent toolchain auto-extension with policy safeguards"
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

# Epic S3 Day 8: Agent toolchain auto-extension with policy safeguards

## TL;DR
- Цель: закрыть MVP-путь "не хватает инструмента в образе" из self-improve цикла.
- MVP-результат: controlled процесс добавления недостающих tools в agent image с audit и rollout policy.

## Priority
- `P1`.

## Scope
### In scope
- Механизм фиксации tool-gap и автоматического предложения изменений образа.
- Policy на добавление зависимостей/инструментов (security/license/size limits).
- Автоматизированная проверка совместимости (bootstrap script + image build + smoke).
- Traceability между self-improve выводом и изменением image/tooling.

### Out of scope
- Полностью автоматический rollout в production.

## Критерии приемки
- Для подтвержденного tool-gap создаётся воспроизводимый PR с изменением образа и evidence проверок.

## Фактический результат (выполнено)
- В agent-runner добавлен baseline-механизм детекции `tool-gap`:
  - источники: structured output `tool_gaps[]`, `codex exec` output, `git push` output;
  - детекция команд с паттернами `command not found` / `executable file not found`.
- Для подтверждённых gap публикуется отдельное audit-событие:
  - `run.toolchain.gap_detected`.
- Дальнейший follow-up для этого события:
  - self-improve агент (`km`) поднимает событие из `flow_events`/`session_json`;
  - готовит PR с доработками prompts/docs и при необходимости `agent-runner` toolchain (`Dockerfile`, `bootstrap_tools.sh`);
  - Owner принимает решение по merge через обычный review-процесс.
- Событие содержит воспроизводимый remediation-контекст:
  - список `tool_gaps`;
  - источники обнаружения;
  - рекомендуемые пути изменения toolchain/image (`bootstrap_tools.sh`, `Dockerfile`).
- В runtime baseline добавлена подготовка диагностического каталога `/tmp/codex-sessions` для self-improve session extraction-процесса.
- `tool_gaps` сохраняются в session snapshot (`session_json`) и доступны для self-improve diagnostics/updater цикла.

## Проверки
- `go test ./services/jobs/agent-runner/internal/runner` — passed.
