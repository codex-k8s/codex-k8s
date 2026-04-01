---
doc_id: IDX-CK8S-ARCH-S18-0001
type: initiative-index
title: "Initiative Package: s18_mission_control_frontend_first_canvas"
status: in-review
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-04-01
related_issues: [480, 561, 562, 563, 565, 567, 571, 573, 579]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-04-01-issue-573-design-package"
---

# s18_mission_control_frontend_first_canvas

## TL;DR
- Пакет объединяет Day4 architecture и Day5 design-артефакты Sprint S18 для frontend-first Mission Control canvas UX на fake data.
- Внутри зафиксированы service boundaries, frontend-only source/state contracts, fake-data data model, workflow preview semantics и explicit handover seam к backend rebuild `#563`.
- Repo-seed prompts остаются source of truth; workflow preview materializes only as deterministic generated block with source refs.
- Следующий обязательный этап после review этого пакета: owner-managed issue `#579` для `run:plan`, где design baseline должен быть разложен в execution package без reopening Sprint S18 baseline.

## Содержимое
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/README.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/architecture.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/c4_context.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/c4_container.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/design_doc.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/api_contract.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/data_model.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/migrations_policy.md`
- `docs/architecture/adr/ADR-0018-mission-control-frontend-first-prototype-and-backend-handover-boundary.md`
- `docs/architecture/alternatives/ALT-0010-mission-control-frontend-first-prototype-boundaries.md`

## Связанные source-of-truth документы
- `docs/delivery/epics/s18/epic-s18-day4-mission-control-frontend-first-canvas-arch.md`
- `docs/delivery/epics/s18/epic-s18-day5-mission-control-frontend-first-canvas-design.md`
- `docs/delivery/epics/s18/epic-s18-day3-mission-control-frontend-first-canvas-prd.md`
- `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md`
- `docs/delivery/sprints/s18/sprint_s18_mission_control_frontend_first_canvas_fake_data.md`
- `docs/delivery/traceability/s18_mission_control_frontend_first_canvas_history.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/architecture/prompt_templates_policy.md`
- `services/staff/web-console/README.md`
- `services/external/api-gateway/README.md`
- `services/internal/control-plane/README.md`
- `services/jobs/worker/README.md`
- `services/jobs/agent-runner/README.md`

## Continuity after `run:design`
- Документный контур `intake -> vision -> prd -> arch -> design` согласован и доведён до review-ready package.
- Day4 закрепил:
  - `web-console` как owner isolated fake-data prototype и локального canvas/view-state;
  - `api-gateway` как thin-edge boundary без новой Mission Control доменной логики;
  - `control-plane` и `worker` как deferred handover owners для backend rebuild `#563`, а не как hidden prerequisite Sprint S18;
  - repo-seed prompts как source of truth и `workflow-policy block` как единственно допустимую форму workflow preview semantics;
  - отсутствие live provider mutation path, DB prompt editor и любых обязательных backend/runtime prerequisites для Sprint S18 `run:dev`.
- Day5 (`#573`) добавил:
  - feature-local async-friendly source contract для fake-data prototype;
  - explicit UI/view-state model для fullscreen canvas, drawer и workflow preview;
  - no-op migration policy: без OpenAPI/proto/schema/runtime changes;
  - documented replacement seam к backend rebuild `#563`.
- Следующий этап `run:plan` идёт через issue `#579` и должен разложить package в execution waves, сохранив continuity `plan -> dev` без разрывов.
