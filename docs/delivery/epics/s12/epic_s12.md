---
doc_id: EPC-CK8S-0012
type: epic
title: "Epic Catalog: Sprint S12 (GitHub API rate-limit resilience, wait-state UX and MCP backpressure)"
status: completed
owner_role: PM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-366-intake"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic Catalog: Sprint S12 (GitHub API rate-limit resilience, wait-state UX and MCP backpressure)

## TL;DR
- Sprint S12 открывает отдельную cross-cutting инициативу вокруг GitHub API rate-limit resilience: платформа должна предсказуемо пережидать budget exhaustion и показывать пользователю, что именно происходит.
- Day1 intake (`#366`) фиксирует проблему, MVP scope, guardrails и continuity в `run:vision`.
- Day2 vision (`#413`) фиксирует mission, persona outcomes, KPI/guardrails и handover в PRD continuity issue `#416`.
- Day3 PRD (`#416`) формализует user stories, FR/AC/NFR, edge cases и expected evidence и передаёт continuity в architecture issue `#418`.
- Day4 architecture (`#418`) закрепляет ownership matrix и handover в design issue `#420`.
- Day5 design (`#420`) фиксирует typed wait contracts, persisted model, finite auto-resume policy и continuity issue `#423` для `run:plan`.
- Day6 plan (`#423`) фиксирует execution package `#425..#431`, sequencing-waves, quality-gates и owner-managed handover в `run:dev`.
- Документный контур `intake -> vision -> prd -> arch -> design -> plan` согласован и завершён; trigger-лейблы на execution waves остаются owner-managed.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s12/epic-s12-day1-github-api-rate-limit-intake.md` (Issue `#366`).
- Day 2 (Vision): `docs/delivery/epics/s12/epic-s12-day2-github-api-rate-limit-vision.md` (Issue `#413`).
- Day 3 (PRD): `docs/delivery/epics/s12/epic-s12-day3-github-api-rate-limit-prd.md` + `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md` (Issue `#416`).
- Day 4 (Architecture): `docs/delivery/epics/s12/epic-s12-day4-github-api-rate-limit-arch.md` + architecture initiative package (Issue `#418`).
- Day 5 (Design): `docs/delivery/epics/s12/epic-s12-day5-github-api-rate-limit-design.md` + design initiative package (Issue `#420`).
- Day 6 (Plan): `docs/delivery/epics/s12/epic-s12-day6-github-api-rate-limit-plan.md` (Issue `#423`); execution package зафиксирован через issues `#425..#431`.

## Execution package (S12-E01..S12-E07)

| Stream | Implementation issue | Scope | Почему это отдельный поток |
|---|---|---|---|
| `S12-E01` | `#425` | Wait-state persistence, additive schema, evidence ledger, dominant wait linkage | Без foundation-слоя остальные streams потеряют единый persisted source-of-truth |
| `S12-E02` | `#426` | `control-plane` classification, contour attribution, visibility projection, resume policy | Domain semantics должны остаться в одном owner-сервисе до старта orchestration и visibility surfaces |
| `S12-E03` | `#427` | `worker` auto-resume sweeps, bounded retry attempts, manual-action escalation | Time-based reconciliation нельзя смешивать с domain classification или runner transport |
| `S12-E04` | `#428` | `agent-runner` signal handoff, session snapshot persistence, deterministic resume payload | Agent path требует отдельной проверки no-local-retry discipline и typed resume contract |
| `S12-E05` | `#429` | Contract-first `api-gateway` visibility routes, DTO/casters, realtime envelopes | Edge должен остаться thin-edge и не принимать доменные решения о rate-limit semantics |
| `S12-E06` | `#430` | `web-console` wait queue, run visibility, contour attribution, manual-action guidance | UX-прозрачность должна идти только поверх typed API, без UI drift и log parsing |
| `S12-E07` | `#431` | Observability, rollout/rollback readiness, acceptance evidence | Без отдельного evidence gate нельзя безопасно перейти в `run:qa` |

## Delivery-governance правила
- До `run:plan` Sprint S12 не создаёт implementation issues и не добавляет runtime/library decisions в репозиторий.
- Plan stage создал implementation issues `#425..#431` без trigger-лейблов; их запуск остаётся owner-managed.
- Каждый stage создаёт следующую issue без trigger-лейбла; запуск следующего stage остаётся owner-managed.
- Rate-limit resilience рассматривается как единая инициатива только пока сохраняется один product story: controlled wait-state, transparency и безопасный resume для GitHub-first контуров.
- Если на `run:vision` выяснится, что notification/adapters или provider abstraction становятся самостоятельным потоком ценности, они выделяются в отдельный follow-up issue, а не раздувают core Sprint S12.
- После `run:prd` continuity зафиксирована через issue `#418`; после `run:arch` подготовлена issue `#420` для `run:design`; после `run:design` подготовлена issue `#423` для `run:plan`.
- Core `run:dev` sequencing фиксируется как `#425 -> #426 -> #427 -> #428 -> #429 -> #430 -> #431`; массовый параллельный старт execution issues запрещён.
- Handover в `run:qa` допускается только после закрытия `#431` и подтверждённого observability/readiness evidence.
