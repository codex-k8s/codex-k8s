---
doc_id: EPC-CK8S-0009
type: epic
title: "Epic Catalog: Sprint S9 (Mission Control Dashboard and console control plane)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335]
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
- До `run:plan` Sprint S9 остаётся markdown-only контуром: execution issues и library lock-in откладываются до подтверждённых product/architecture decisions.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s9/epic-s9-day1-mission-control-dashboard-intake.md` (Issue `#333`).
- Day 2 (Vision): follow-up issue `#335` (`run:vision`) после owner review intake-пакета.
- Day 3 (PRD): issue создаётся на этапе vision; должна формализовать user stories, FR/AC/NFR и waves.
- Day 4 (Architecture): issue создаётся на этапе PRD; должна зафиксировать service boundaries, projections, realtime contracts и provider-sync ownership.
- Day 5 (Design): issue создаётся на этапе architecture; должна подготовить implementation-ready package по API/data/UI/reconciliation.
- Day 6 (Plan): issue создаётся на этапе design; должна сформировать execution backlog и quality-gates перед `run:dev`.

## Candidate product streams

| Epic ID | Scope | Почему это отдельный поток |
|---|---|---|
| `S9-E01` | Active-set dashboard shell: landing page, summary, board/list toggle, side panel | Это core UX-слой, который должен давать situational awareness за 5-10 секунд |
| `S9-E02` | Work item / discussion / PR / agent модель и связи | Без общей сущностной модели dashboard превратится в набор несвязанных карточек |
| `S9-E03` | Realtime snapshot/delta, side-panel chat, comment/timeline projections | Realtime и chat формируют основной operational loop и требуют отдельной data/reconciliation проработки |
| `S9-E04` | UI-command -> outbound sync -> webhook echo dedupe/correlation | Это ключевой guardrail против дублей issue/run/comment и ложных повторных запусков |
| `S9-E05` | Voice intake и AI-assisted draft structuring | Высокая ценность для discussion-first сценария, но отдельный риск/зависимость по AI policy и ROI |

## Delivery-governance правила
- До `run:plan` Sprint S9 не создаёт implementation issues и не добавляет внешние зависимости в репозиторий.
- Каждый stage обязан создавать следующую issue без trigger-лейбла; trigger на запуск следующего stage ставит Owner.
- `S9-E05` не считается blocking scope для первой волны dashboard MVP, пока vision/PRD не докажут его приоритет и измеримый вклад.
