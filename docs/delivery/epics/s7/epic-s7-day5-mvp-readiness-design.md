---
doc_id: EPC-CK8S-S7-D5
type: epic
title: "Epic S7 Day 5: Design для закрытия MVP readiness gaps (Issue #238)"
status: in-review
owner_role: SA
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [212, 218, 220, 222, 238, 241]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-238-design-epic"
---

# Epic S7 Day 5: Design для закрытия MVP readiness gaps (Issue #238)

## TL;DR
- Подготовлен design package для перехода в `run:plan`:
  - `design_doc`,
  - `api_contract`,
  - `data_model`,
  - `migrations_policy`.
- Для потоков `S7-E06/S7-E07/S7-E09/S7-E10/S7-E13/S7-E16/S7-E17` зафиксированы typed contract decisions и risk-mitigation.
- Для потоков с persisted-state impact зафиксированы migration/rollback правила и rollout sequence.

## Контекст
- Stage continuity: `#212 -> #218 -> #220 -> #222 -> #238 -> #241`.
- Входной baseline: architecture package Day4 (`ADR-0010`, `ALT-0002`, S7 ownership matrix).
- Scope Day5: markdown-only, без runtime/code изменений.

## Артефакты Day5
- `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/design_doc.md`
- `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/api_contract.md`
- `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/data_model.md`
- `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/migrations_policy.md`

## Ключевые design-решения
- `S7-E06`: контракт agents settings де-scoped до MVP (runtime mode/locale удалены из write DTO).
- `S7-E07`: prompt source зафиксирован как repo-only без selector `repo|db`.
- `S7-E09`: namespace delete path остаётся через существующий typed endpoint, UI cleanup без contract drift.
- `S7-E10`: добавлены typed action decisions для cancel/stop runtime deploy tasks с idempotency.
- `S7-E13`: stage resolver расширяется late-stage revise coverage `run:doc-audit|qa|release|postdeploy|ops|self-improve:revise` с deterministic ambiguity handling.
- `S7-E16`: run finalization нормализуется typed terminal metadata для исключения false-failed.
- `S7-E17`: snapshot rewrite reliability через version/checksum contract и CAS-like upsert semantics.

## Context7 verification
- Проверен baseline `kin-openapi` (`/getkin/kin-openapi`) для request/response validation path.
- Проверен baseline `monaco-editor` (`/microsoft/monaco-editor`) для diff editor API (`createDiffEditor`/`setModel`).
- Новые внешние зависимости в Day5 не требуются.

## Runtime и migration impact
- Runtime impact на этапе Day5: отсутствует (markdown-only).
- Migration impact для `run:dev`:
  - additive columns/indexes в `runtime_deploy_tasks`, `agent_runs`, `agent_sessions`;
  - event payload hardening в `flow_events`;
  - optional prompt-source normalization для `prompt_templates`.
- Rollout order остаётся обязательным:
  `migrations -> internal -> edge -> frontend`.

## Acceptance criteria status (Issue #238)
- [x] Подготовлен design package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`) по потокам с impact.
- [x] Для `S7-E06/S7-E07/S7-E09/S7-E10/S7-E13/S7-E16/S7-E17` зафиксированы typed contract decisions и риск-mitigation.
- [x] Для потоков с persisted-state изменениями зафиксированы migration/rollback правила.
- [x] Синхронизированы `issue_map`, `requirements_traceability`, sprint/epic docs.
- [x] Подготовлена следующая stage-issue `run:plan` без trigger-лейбла (`#241`).

## Handover в `run:plan`
- Следующий этап: `run:plan`.
- Follow-up issue: `#241` (без trigger-лейбла; trigger ставит Owner).
- Обязательная декомпозиция в `run:plan`:
  - execution issues по contract/data/runtime потокам;
  - quality-gates для parity-gate перед `run:dev`;
  - sequencing constraints для rollout/review/qa.
