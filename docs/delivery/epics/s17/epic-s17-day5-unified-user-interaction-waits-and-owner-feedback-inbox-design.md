---
doc_id: EPC-CK8S-S17-D5-OWNER-FEEDBACK-DESIGN
type: epic
title: "Epic S17 Day 5: Design для unified owner feedback loop (Issues #568/#575)"
status: in-review
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [541, 554, 557, 559, 568, 575]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-568-design-epic"
---

# Epic S17 Day 5: Design для unified owner feedback loop (Issues #568/#575)

## TL;DR
- Подготовлен полный Day5 design package Sprint S17 для unified owner feedback loop:
  - `design_doc`,
  - `api_contract`,
  - `data_model`,
  - `migrations_policy`.
- Зафиксированы typed contracts для built-in wait path, Telegram callbacks, staff-console fallback, response binding registry, wait-state linkage и recovery-only snapshot resume.
- Сохранены Day4 boundaries: same-session live continuation как primary happy-path, timeout/TTL baseline не ниже owner wait window, Telegram + staff-console как thin surfaces поверх одного persisted truth.
- Создана follow-up issue `#575` для stage `run:plan` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#541`.
- Vision baseline: `#554`.
- PRD baseline: `#557`.
- Architecture baseline: `#559`.
- Текущий этап: `run:design` в Issue `#568`.
- Scope этапа: только markdown-изменения.

## Design package
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/README.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/design_doc.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/api_contract.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/data_model.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/migrations_policy.md`
- `docs/delivery/traceability/s17_unified_user_interaction_waits_and_owner_feedback_inbox_history.md`

## Ключевые design-решения
- Built-in wait entrypoint stays on `user.decision.request`; control tool `owner.feedback.request` is not reused.
- Owner-feedback request truth remains platform-owned in `control-plane` and extends Sprint S10/S11 interaction foundation additively.
- Staff-console is modeled as projection + typed response surface, not as a second adapter owner.
- One response binding registry classifies Telegram callback/free-text/voice and staff-console responses into the same winner-selection path.
- Recovery resume remains explicit degraded classification and never rewrites history into normal happy-path.

## Acceptance Criteria (Issue #568)
- [x] Подготовлен design package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`).
- [x] Зафиксированы typed contracts, data model, migration/rollout notes и visibility/fallback rules без пересмотра Day4 boundaries.
- [x] Сохранены blocking baselines: same-session happy-path, timeout/TTL baseline, recovery-only snapshot-resume, dual-channel truth и `run:self-improve` exclusion.
- [x] Staff-console fallback не стал вторым source of truth, а Telegram не стал владельцем platform semantics.
- [x] Подготовлена follow-up issue `#575` для stage `run:plan`.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S17-D5-01` Contract completeness | Есть `design_doc + api_contract + data_model + migrations_policy` | passed |
| `QG-S17-D5-02` Same-session integrity | Wait linkage and continuation path сохраняют live same-session happy-path как primary model | passed |
| `QG-S17-D5-03` Single persisted truth | Telegram and staff-console read/write through one aggregate + projection model | passed |
| `QG-S17-D5-04` Recovery discipline | Recovery resume классифицируется отдельно и не маскирует runtime loss | passed |
| `QG-S17-D5-05` Rollout discipline | Зафиксированы additive migrations и rollout order относительно Sprint S10/S11 | passed |
| `QG-S17-D5-06` Stage continuity | Создана issue `#575` на `run:plan` без trigger-лейбла | passed |

## Handover в `run:plan`
- Следующий этап: `run:plan`.
- Follow-up issue: `#575`.
- Trigger-лейбл на новую issue ставит Owner после review design package.
- На plan-stage обязательно:
  - декомпозировать execution waves по schema, domain, worker reconcile, edge transport, Telegram adapter и web-console;
  - зафиксировать quality gates, DoR/DoD и acceptance evidence для same-session / recovery / manual-fallback paths;
  - продолжить issue-цепочку `plan -> dev` без разрывов.
