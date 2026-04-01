---
doc_id: EPC-CK8S-0018
type: epic
title: "Epic Catalog: Sprint S18 (Frontend-first Mission Control canvas UX on fake data)"
status: in-review
owner_role: PM
created_at: 2026-03-26
updated_at: 2026-04-01
related_issues: [470, 480, 522, 523, 524, 525, 561, 562, 563, 565, 567, 571, 573, 579]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-26-issue-562-intake"
---

# Epic Catalog: Sprint S18 (Frontend-first Mission Control canvas UX on fake data)

## TL;DR
- Sprint S18 открывает отдельный Mission Control reset-stream после doc-reset `#561`: сначала owner утверждает frontend-first UX на fake data, затем отдельной задачей запускается backend rebuild `#563`.
- Day1 intake (`#562`) зафиксировал новый baseline: fullscreen свободный canvas, node taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls и workflow editor UX без live GitHub mutation path.
- Day2 vision (`#565`) закрепил mission, north star, persona outcomes, KPI/guardrails и wave boundaries для canvas-first UX и создал issue `#567` для `run:prd`.
- Day3 PRD (`#567`) формализовал user stories, FR/AC/NFR, scenario matrix и expected evidence для owner walkthrough, operator navigation и workflow policy preview на fake data, а также создал issue `#571` для `run:arch`.
- Day4 architecture (`#571`) зафиксировал `web-console` как owner isolated fake-data prototype, explicit handover seam к backend rebuild `#563` и создал issue `#573` для `run:design`.
- Day5 design (`#573`) зафиксировал frontend-only source/state contracts, fake-data data model, deterministic workflow preview и создал issue `#579` для `run:plan`.
- Prompt policy не переоткрывается: repo-seed prompts остаются source of truth, а workflow behavior допускается только через deterministic generated `workflow-policy block`.
- До `run:dev` Sprint S18 остаётся stage-driven frontend-first инициативой; на `run:dev` целевой результат ограничен isolated fake-data prototype в `web-console`.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s18/epic-s18-day1-mission-control-frontend-first-canvas-intake.md` (Issue `#562`); stage сформулировал problem statement, reset baseline, scope/guardrails и создал issue `#565` для `run:vision`.
- Day 2 (Vision): `docs/delivery/epics/s18/epic-s18-day2-mission-control-frontend-first-canvas-vision.md` (Issue `#565`); stage зафиксировал mission, north star, persona outcomes, KPI/guardrails и wave boundaries без reopening Day1 baseline и создал issue `#567` для `run:prd`.
- Day 3 (PRD): `docs/delivery/epics/s18/epic-s18-day3-mission-control-frontend-first-canvas-prd.md` + `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md` (Issue `#567`); stage зафиксировал product contract Sprint S18 и создал issue `#571` для `run:arch`.
- Day 4 (Architecture): `docs/delivery/epics/s18/epic-s18-day4-mission-control-frontend-first-canvas-arch.md` + `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/*` + `ADR-0018` + `ALT-0010` (Issue `#571`); stage закрепил ownership split для isolated prototype и отдельный backend handover.
- Day 5 (Design): `docs/delivery/epics/s18/epic-s18-day5-mission-control-frontend-first-canvas-design.md` + `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/{README.md,design_doc.md,api_contract.md,data_model.md,migrations_policy.md}` (Issue `#573`); stage выпустил implementation-ready frontend-only package и создал issue `#579` для `run:plan`.
- Day 6 (Plan): Issue `#579`; создаётся последовательно после design и должна разложить execution package, feedback loops, DoR/DoD и owner-managed handover в `run:dev`.
- Day 7 (Development): owner-managed `run:dev` должен реализовать isolated `web-console` prototype на fake data и завершить именно frontend-first инициативу без обязательного auto-continue в late stages.

## Delivery-governance правила
- Sprint S18 идёт полной цепочкой `intake -> vision -> prd -> arch -> design -> plan -> dev`.
- Каждый doc-stage обязан создавать следующую follow-up issue без trigger-лейбла; запуск следующего stage остаётся owner-managed.
- До `run:dev` Sprint S18 остаётся markdown-only контуром и не создаёт code/runtime diff вне документации.
- Day1 baseline обязателен для всех следующих stage:
  - UX сначала утверждается на fake data, backend rebuild идёт потом в issue `#563`;
  - fullscreen свободный canvas без lane/column shell и без обязательной nested root-group модели;
  - Wave 1 taxonomy `Issue`, `PR`, `Run`;
  - compact nodes, explicit node-to-node relations, side panel/drawer и toolbar/controls обязательны;
  - workflow editor UX работает только на fake data и не становится live mutation path;
  - repo-seed prompts остаются source of truth, free-form DB prompt editor не вводится;
  - `run:dev` ограничен isolated fake-data prototype в `web-console`;
  - автоматическое продолжение в `run:qa -> run:release -> run:postdeploy -> run:ops` внутри этого спринта не требуется.
- Sequencing из rethink `#561` остаётся актуальным:
  - `#522` и `#523` можно продолжать независимо;
  - `#524` и `#525` остаются заблокированными до owner approval результата Sprint S18;
  - `#470` может двигаться только без фиксации финального cockpit UI.
- После design stage следующий owner-managed handover = issue `#579` для `run:plan`; continuity `plan -> dev` должна сохраняться без разрывов.
