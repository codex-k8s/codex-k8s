---
doc_id: EPC-CK8S-0006
type: epic
title: "Epic Catalog: Sprint S6 (Agents configuration and prompt templates lifecycle)"
status: completed
owner_role: PM
created_at: 2026-02-25
updated_at: 2026-03-02
related_issues: [184, 185, 187, 189, 195, 197, 199, 201, 216, 262, 263, 265]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-187-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic Catalog: Sprint S6 (Agents configuration and prompt templates lifecycle)

## TL;DR
- Sprint S6 ведет инициативу по переводу раздела `Agents` из scaffold в production-ready lifecycle контур.
- Day1 intake (`#184`) зафиксировал проблему и границы MVP.
- Day2 vision (`#185`) зафиксировал mission/KPI и риск-рамку.
- Day3 PRD (`#187`) формализовал FR/AC/NFR-draft, итоговый PRD-пакет смержен в `main` через PR `#190`.
- Day4 architecture (`#189`) зафиксировал архитектурный пакет и создал follow-up issue `#195` для `run:design`.
- Day5 design (`#195`) зафиксировал implementation-ready design package и создал follow-up issue `#197` для `run:plan`.
- Day6 plan (`#197`) зафиксировал execution roadmap + quality-gates + DoD и создал follow-up issue `#199` для `run:dev`.
- Day7 dev (`#199`) реализовал lifecycle `agents/templates/audit` (PR `#202`) и создал follow-up issue `#201` для `run:qa`.
- Day8 QA (`#201`) подтвердил readiness и оформил переход в `run:release` через issue `#216`.
- Day9 release closeout (`#262`) завершил release-governance Sprint S6 и создал issue `#263` для `run:postdeploy`.
- Day10 postdeploy (`#263`) зафиксировал runtime evidence, обновил ops handover и подготовил переход в `run:ops`.

## Эпики Sprint S6
- Day 1 (Intake): `docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`
- Day 2 (Vision baseline): `docs/delivery/epics/s6/epic-s6-day2-agents-prompts-vision.md` (Issue `#185`).
- Day 3 (PRD):
  - `docs/delivery/epics/s6/epic-s6-day3-agents-prompts-prd.md`
  - `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md`
- Day 4 (Architecture): `docs/delivery/epics/s6/epic-s6-day4-agents-prompts-arch.md` (Issue `#189`).
- Day 5 (Design): `docs/delivery/epics/s6/epic-s6-day5-agents-prompts-design.md` (Issue `#195`).
- Day 6 (Plan): `docs/delivery/epics/s6/epic-s6-day6-agents-prompts-plan.md` (Issue `#197`).
- Day 9 (Release closeout): `docs/delivery/epics/s6/epic-s6-day9-release-closeout.md` (Issue `#262`).
- Day 10 (Postdeploy review): `docs/delivery/epics/s6/epic-s6-day10-postdeploy-review.md` (Issue `#263`).

## Закрывающие этапы и continuity
- Day 7 (Dev, completed): Issue `#199`, PR `#202`.
- Day 8 (QA, completed): Issue `#201`, решение GO в `run:release`.
- Day 9 (Release, completed): Issue `#262`, release closeout и traceability sync.
- Day 10 (Postdeploy, in-review): Issue `#263`, сформирован postdeploy evidence package и ops handover.
- Следующий этап после `run:postdeploy`: создана issue `#265` для stage `run:ops`.

## Delivery-governance правила
- Каждый stage завершает работу созданием issue для следующего stage.
- Follow-up issue создаются без `run:*`-лейбла; trigger-лейбл на запуск следующего stage ставит Owner.
- Каждая следующая issue обязана содержать явную инструкцию создать issue после завершения текущего этапа.
- Для цепочки S6 зафиксирована последовательность continuity:
  - `#184 (intake) -> #185 (vision) -> #187 (prd) -> #189 (arch) -> #195 (design) -> #197 (plan) -> #199 (dev) -> #201 (qa) -> #216 (release continuity) -> #262 (release closeout) -> #263 (postdeploy) -> #265 (ops)`.
- Для `run:postdeploy` continuity-правило выполнено: подготовлена отдельная issue `run:ops`.
