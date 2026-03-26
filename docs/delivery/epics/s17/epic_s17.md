---
doc_id: EPC-CK8S-0017
type: epic
title: "Epic Catalog: Sprint S17 (Unified long-lived user interaction waits and owner feedback inbox)"
status: in-review
owner_role: PM
created_at: 2026-03-20
updated_at: 2026-03-26
related_issues: [360, 361, 458, 473, 532, 540, 541, 554, 557, 559, 568]
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
- Day2 vision package (`#554`) закрепил mission, persona outcomes, KPI/guardrails, max timeout/TTL baseline для built-in `codex_k8s` MCP wait path и wave boundaries без переоткрытия Day1 baseline и создал issue `#557` для `run:prd`.
- Day3 PRD package (`#557`) зафиксировал user stories, FR/AC/NFR, scenario matrix, expected evidence, recovery/lifecycle guardrails и создал issue `#559` для `run:arch`.
- Day4 architecture package (`#559`) зафиксировал service boundaries, ownership split, live wait lifetime policy, persisted request truth и recovery-only snapshot-resume boundary и создал issue `#568` для `run:design`.
- До `run:plan` Sprint S17 остаётся markdown-only контуром: кодовые/runtime changes и конкретные schema/API decisions начинаются только после owner review следующих stage.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md` (Issue `#541`).
- Day 2 (Vision): `docs/delivery/epics/s17/epic-s17-day2-unified-user-interaction-waits-and-owner-feedback-inbox-vision.md` (Issue `#554`); stage зафиксировал mission, persona outcomes, KPI/guardrails, max timeout/TTL baseline для built-in `codex_k8s` MCP wait path и wave boundaries для owner feedback loop и создал issue `#557` для `run:prd`.
- Day 3 (PRD): issue `#557`, `docs/delivery/epics/s17/epic-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox-prd.md` и `docs/delivery/epics/s17/prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`; stage формализовал user stories, FR/AC/NFR, scenario matrix и expected evidence и создал issue `#559` для `run:arch`.
- Day 4 (Architecture): issue `#559`, `docs/delivery/epics/s17/epic-s17-day4-unified-user-interaction-waits-and-owner-feedback-inbox-arch.md` и package `docs/architecture/initiatives/s17_unified_owner_feedback_loop/*`; stage зафиксировал execution model, ownership split, lifetime policy, persisted truth и создал issue `#568` для `run:design`.
- Day 5 (Design): issue `#568`, создаётся последовательно после architecture и должна выпустить implementation-ready API/data/UI/runtime contract.
- Day 6 (Plan): создаётся последовательно после design и должна разложить execution package, quality-gates и owner-managed handover в `run:dev`.

## Delivery-governance правила
- Sprint S17 идёт полным doc-stage контуром `intake -> vision -> prd -> arch -> design -> plan`.
- Каждый stage обязан создавать следующую follow-up issue без trigger-лейбла; запуск следующего stage остаётся owner-managed.
- До `run:plan` Sprint S17 не создаёт code/runtime changes и не фиксирует premature implementation details.
- Day1 baseline обязателен для всех следующих stage:
  - same live pod / same `codex` session как primary happy-path;
  - built-in `codex_k8s` MCP wait path обязан использовать max timeout/TTL не ниже owner wait window;
  - persisted session snapshot только как recovery fallback;
  - long human-wait target не меньше 24 часов;
  - lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`;
  - Telegram pending inbox + staff-console fallback на одном persisted backend contract;
  - deterministic text/voice binding;
  - `run:self-improve` вне human-wait contract.
- Day4 architecture baseline дополнительно обязателен для `run:design`:
  - `control-plane` остаётся owner persisted request truth и accepted-response winner;
  - `worker` владеет dispatch/reconcile/lease keepalive, а `agent-runner` только live session + recovery snapshot;
  - `api-gateway`, `staff web-console` и `telegram-interaction-adapter` остаются thin surfaces.
- Detached resume-run не может вернуться как default UX без нового owner-решения.
