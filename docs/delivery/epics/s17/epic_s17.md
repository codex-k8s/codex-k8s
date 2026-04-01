---
doc_id: EPC-CK8S-0017
type: epic
title: "Epic Catalog: Sprint S17 (Unified long-lived user interaction waits and owner feedback inbox)"
status: in-review
owner_role: PM
created_at: 2026-03-20
updated_at: 2026-04-01
related_issues: [360, 361, 458, 473, 532, 540, 541, 554, 557, 559, 568, 575, 582]
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
- Day5 design package (`#568`) зафиксировал implementation-ready API/data/runtime contracts, response binding registry, staff-console fallback surface и additive migration policy поверх Sprint S10/S11 foundation и создал issue `#575` для `run:plan`.
- Day6 plan package (`#575`) зафиксировал execution package `S17-E01..S17-E07`, prerequisite gate на закрытых Sprint S10/S11 foundation issues `#391..#395` и `#458`, quality-gates, DoR/DoD и создал handover issue `#582` для `run:dev`.
- До owner-managed запуска `run:dev` в issue `#582` Sprint S17 остаётся markdown-only контуром: кодовые/runtime changes и конкретные schema/API decisions начинаются только после review plan package.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md` (Issue `#541`).
- Day 2 (Vision): `docs/delivery/epics/s17/epic-s17-day2-unified-user-interaction-waits-and-owner-feedback-inbox-vision.md` (Issue `#554`); stage зафиксировал mission, persona outcomes, KPI/guardrails, max timeout/TTL baseline для built-in `codex_k8s` MCP wait path и wave boundaries для owner feedback loop и создал issue `#557` для `run:prd`.
- Day 3 (PRD): issue `#557`, `docs/delivery/epics/s17/epic-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox-prd.md` и `docs/delivery/epics/s17/prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`; stage формализовал user stories, FR/AC/NFR, scenario matrix и expected evidence и создал issue `#559` для `run:arch`.
- Day 4 (Architecture): issue `#559`, `docs/delivery/epics/s17/epic-s17-day4-unified-user-interaction-waits-and-owner-feedback-inbox-arch.md` и package `docs/architecture/initiatives/s17_unified_owner_feedback_loop/*`; stage зафиксировал execution model, ownership split, lifetime policy, persisted truth и создал issue `#568` для `run:design`.
- Day 5 (Design): issue `#568`, `docs/delivery/epics/s17/epic-s17-day5-unified-user-interaction-waits-and-owner-feedback-inbox-design.md` и `docs/architecture/initiatives/s17_unified_owner_feedback_loop/{design_doc.md,api_contract.md,data_model.md,migrations_policy.md}`; stage выпустил implementation-ready API/data/UI/runtime contract и создал issue `#575` для `run:plan`.
- Day 6 (Plan): issue `#575`, `docs/delivery/epics/s17/epic-s17-day6-unified-user-interaction-waits-and-owner-feedback-inbox-plan.md`; stage выпустил execution package `S17-E01..S17-E07`, quality-gates, DoR/DoD и handover issue `#582` для owner-managed запуска `run:dev`.

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
- Day5 design baseline дополнительно обязателен для `run:plan`:
  - `user.decision.request` остаётся built-in wait entrypoint, а `owner.feedback.request` не переиспользуется;
  - response binding registry классифицирует Telegram callback/free-text/voice и staff-console responses в один winner-selection path;
  - staff-console моделируется как projection + typed response surface, а не как второй source of truth;
  - recovery resume остаётся explicit degraded path.
- Day6 plan baseline дополнительно обязателен для `run:dev`:
  - единый execution anchor = issue `#582`, скрытый параллельный dev-backlog не допускается;
  - waves `S17-E01 -> S17-E02 -> S17-E03 -> S17-E04 -> S17-E05 -> S17-E06 -> S17-E07` обязательны как sequencing contract;
  - prerequisite gate на закрытых Sprint S10/S11 foundation issues `#391..#395` и `#458` сохраняется до старта кода;
  - observability/evidence gate остаётся обязательным перед `run:qa`.
- Detached resume-run не может вернуться как default UX без нового owner-решения.
