---
doc_id: EPC-CK8S-0012
type: epic
title: "Epic Catalog: Sprint S12 (GitHub API rate-limit resilience, wait-state UX and MCP backpressure)"
status: in-review
owner_role: PM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-issue-366-intake"
---

# Epic Catalog: Sprint S12 (GitHub API rate-limit resilience, wait-state UX and MCP backpressure)

## TL;DR
- Sprint S12 открывает отдельную cross-cutting инициативу вокруг GitHub API rate-limit resilience: платформа должна предсказуемо пережидать budget exhaustion и показывать пользователю, что именно происходит.
- Day1 intake (`#366`) фиксирует проблему, MVP scope, guardrails и continuity в `run:vision`.
- Day2 vision (`#413`) фиксирует mission, persona outcomes, KPI/guardrails и handover в PRD continuity issue `#416`.
- Day3 PRD (`#416`) формализует user stories, FR/AC/NFR, edge cases и expected evidence и передаёт continuity в architecture issue `#418`.
- Дальнейшие stage-issues (`arch -> design -> plan`) создаются последовательно после review предыдущего этапа; trigger-лейблы остаются owner-managed.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s12/epic-s12-day1-github-api-rate-limit-intake.md` (Issue `#366`).
- Day 2 (Vision): `docs/delivery/epics/s12/epic-s12-day2-github-api-rate-limit-vision.md` (Issue `#413`).
- Day 3 (PRD): `docs/delivery/epics/s12/epic-s12-day3-github-api-rate-limit-prd.md` + `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md` (Issue `#416`).
- Day 4 (Architecture): `docs/delivery/epics/s12/epic-s12-day4-github-api-rate-limit-arch.md` + architecture initiative package (Issue `#418`).
- Day 5 (Design): Issue `#420`, создаётся на выходе `run:arch`.
- Day 6 (Plan): TBD, создаётся на выходе `run:design`.

## Delivery-governance правила
- До `run:plan` Sprint S12 не создаёт implementation issues и не добавляет runtime/library decisions в репозиторий.
- Каждый stage создаёт следующую issue без trigger-лейбла; запуск следующего stage остаётся owner-managed.
- Rate-limit resilience рассматривается как единая инициатива только пока сохраняется один product story: controlled wait-state, transparency и безопасный resume для GitHub-first контуров.
- Если на `run:vision` выяснится, что notification/adapters или provider abstraction становятся самостоятельным потоком ценности, они выделяются в отдельный follow-up issue, а не раздувают core Sprint S12.
- После `run:prd` continuity зафиксирована через issue `#418`; после `run:arch` подготовлена issue `#420` для `run:design`; дальнейшие `plan` issues создаются только после review предыдущего stage.
