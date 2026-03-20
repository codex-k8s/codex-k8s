---
doc_id: EPC-CK8S-0017
type: epic
title: "Epic Catalog: Sprint S17 (Unified long-lived user interaction waits and owner feedback inbox)"
status: in-review
owner_role: PM
created_at: 2026-03-20
updated_at: 2026-03-20
related_issues: [360, 361, 458, 473, 532, 540, 541, 554]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-20-issue-541-intake"
---

# Epic Catalog: Sprint S17 (Unified long-lived user interaction waits and owner feedback inbox)

## TL;DR
- Sprint S17 открывает отдельную cross-cutting инициативу вокруг long-lived owner feedback loop: built-in user interactions и Telegram channel уже существуют, но всё ещё не дают deterministic contract `delivery -> wait -> response -> same-session continuation`.
- Day1 intake (`#541`) зафиксировал problem statement, hybrid execution model, long human-wait baseline `>=24h`, delivery-before-wait lifecycle, Telegram pending inbox, staff-console fallback и persisted text/voice binding.
- Day2 continuity issue `#554` открывает `run:vision` и должна закрепить mission, persona outcomes, KPI/guardrails и wave boundaries без переоткрытия Day1 baseline.
- До `run:plan` Sprint S17 остаётся markdown-only контуром: кодовые/runtime changes и конкретные schema/API decisions начинаются только после owner review следующих stage.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md` (Issue `#541`).
- Day 2 (Vision): continuity issue `#554` для `run:vision`; stage должен зафиксировать mission, persona outcomes, KPI/guardrails и wave boundaries для owner feedback loop.
- Day 3 (PRD): создаётся последовательно после owner review Day2 и должна формализовать user stories, FR/AC/NFR и expected evidence.
- Day 4 (Architecture): создаётся последовательно после PRD и должна зафиксировать execution model, ownership split и lifetime policy.
- Day 5 (Design): создаётся последовательно после architecture и должна выпустить implementation-ready API/data/UI/runtime contract.
- Day 6 (Plan): создаётся последовательно после design и должна разложить execution package, quality-gates и owner-managed handover в `run:dev`.

## Delivery-governance правила
- Sprint S17 идёт полным doc-stage контуром `intake -> vision -> prd -> arch -> design -> plan`.
- Каждый stage обязан создавать следующую follow-up issue без trigger-лейбла; запуск следующего stage остаётся owner-managed.
- До `run:plan` Sprint S17 не создаёт code/runtime changes и не фиксирует premature implementation details.
- Day1 baseline обязателен для всех следующих stage:
  - same live pod / same `codex` session как primary happy-path;
  - persisted session snapshot только как recovery fallback;
  - long human-wait target не меньше 24 часов;
  - lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`;
  - Telegram pending inbox + staff-console fallback на одном persisted backend contract;
  - deterministic text/voice binding;
  - `run:self-improve` вне human-wait contract.
- Detached resume-run не может вернуться как default UX без нового owner-решения.
