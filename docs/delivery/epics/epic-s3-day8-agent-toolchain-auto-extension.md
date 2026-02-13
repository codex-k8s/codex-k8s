---
doc_id: EPC-CK8S-S3-D8
type: epic
title: "Epic S3 Day 8: Agent toolchain auto-extension with policy safeguards"
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
