---
doc_id: EPC-CK8S-0009
type: epic
title: "Epic Catalog: Sprint S9 (Mission Control Dashboard and console control plane)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-14
related_issues: [333, 335, 337, 340, 351, 363, 369, 370, 371, 372, 373, 374, 375]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-333-intake"
---

# Epic Catalog: Sprint S9 (Mission Control Dashboard and console control plane)

## TL;DR
- Sprint S9 открывает отдельную продуктовую инициативу Mission Control Dashboard: staff console должна стать основным control-plane UX для активной работы, а не только набором runtime/debug страниц.
- Day1 intake (`#333`) фиксирует проблему, границы MVP и continuity в `run:vision`.
- Day2 vision (`#335`) закрепляет mission, persona outcomes, KPI/guardrails и handover в PRD issue `#337`.
- Day3 PRD (`#337`) формализует user stories, FR/AC/NFR, edge cases и wave priorities и передаёт continuity в architecture issue `#340`.
- Day4 architecture (`#340`) закрепляет ownership для projections, commands, provider sync/reconciliation и degraded realtime path и передаёт continuity в design issue `#351`.
- Day5 design (`#351`) фиксирует implementation-ready contracts по API/data/realtime/rollout и передаёт continuity в plan issue `#363`.
- Day6 plan (`#363`) формирует execution backlog `#369..#375`, sequencing-waves, quality-gates и owner decisions для перехода в `run:dev`.
- Owner revision 2026-03-14: stream `S9-E06` / Issue `#374` переведён в superseded historical artifact и выведен из активного execution backlog; отдельная observability/readiness wave для Sprint S9 сейчас не планируется.
- Voice contour вынесен в отдельный conditional stream `#375` и не блокирует core dashboard MVP.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s9/epic-s9-day1-mission-control-dashboard-intake.md` (Issue `#333`).
- Day 2 (Vision): `docs/delivery/epics/s9/epic-s9-day2-mission-control-dashboard-vision.md` (Issue `#335`).
- Day 3 (PRD): `docs/delivery/epics/s9/epic-s9-day3-mission-control-dashboard-prd.md` + `docs/delivery/epics/s9/prd-s9-day3-mission-control-dashboard.md` (Issue `#337`).
- Day 4 (Architecture): `docs/delivery/epics/s9/epic-s9-day4-mission-control-dashboard-arch.md` + architecture package in `docs/architecture/initiatives/s9_mission_control_dashboard/` (Issue `#340`).
- Day 5 (Design): `docs/delivery/epics/s9/epic-s9-day5-mission-control-dashboard-design.md` + design package in `docs/architecture/initiatives/s9_mission_control_dashboard/` (Issue `#351`).
- Day 6 (Plan): `docs/delivery/epics/s9/epic-s9-day6-mission-control-dashboard-plan.md` (Issue `#363`) + execution issues `#369..#375`.

## Execution streams (finalized by Issue #363)

| Stream | Implementation issue | Wave | Scope | Почему это отдельный поток |
|---|---:|---|---|---|
| `S9-E01` | `#369` | Wave 1 | Projection schema, additive indexes, repository foundation и rollout guards | Без foundation-stream остальные потоки рискуют стартовать на непроверенной persisted projection |
| `S9-E02` | `#370` | Wave 2 | `control-plane` active-set model, relations и command lifecycle | Это единственный owner persisted projection и admission policy |
| `S9-E03` | `#371` | Wave 3 | `worker` warmup/backfill execution, provider sync/retry и echo dedupe | Нужен отдельный background contour для idempotent provider mutations, rebuild и recovery |
| `S9-E04` | `#372` | Wave 3 | Core contract-first `api-gateway` transport и realtime envelope | Edge должен сохранить thin-edge boundary и transport consistency без смешения с optional voice path |
| `S9-E05` | `#373` | Wave 4 | Dashboard shell, board/list toggle и side panel state integration | Core UX-слой, который даёт situational awareness и explicit degraded fallback |
| `S9-E06` | `#374` | Wave 5 | Superseded 2026-03-14: отдельная observability/readiness wave больше не планируется | Исторический artifact plan-stage; код из PR `#463` Owner не принял |
| `S9-E07` | `#375` | Wave 6 (conditional) | Optional voice-candidate transport + rollout contour | Высокая ценность, но отдельный риск по ROI, policy и dependency choices |

## Delivery-governance правила
- До `run:plan` Sprint S9 не создаёт implementation issues и не добавляет внешние зависимости в репозиторий.
- После `run:plan` созданы handover issues `#369..#375` без trigger-лейблов; запуск `run:dev` остаётся owner-managed по waves.
- Каждый stage обязан создавать следующую issue без trigger-лейбла; trigger на запуск следующего stage ставит Owner.
- Warmup/backfill execution закреплён за `#371`, а `#369` ограничен schema/repository foundation и rollout guards.
- Voice-specific OpenAPI/codegen/DTO/casters закреплены за `#375`; `#372` покрывает только core snapshot/details/commands/realtime transport.
- `S9-E06` / `#374` не считается частью активного core backlog после owner revision 2026-03-14 и сохраняется только как historical artifact.
- `S9-E07` не считается blocking scope для первой волны dashboard MVP, пока active core backlog `#369..#373` не даст подтверждённый value/evidence.
