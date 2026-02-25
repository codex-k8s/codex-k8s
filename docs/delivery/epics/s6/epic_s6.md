---
doc_id: EPC-CK8S-0006
type: epic
title: "Epic Catalog: Sprint S6 (Agents configuration and prompt templates lifecycle)"
status: in-progress
owner_role: PM
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-25-issue-187-prd"
---

# Epic Catalog: Sprint S6 (Agents configuration and prompt templates lifecycle)

## TL;DR
- Sprint S6 ведет инициативу по переводу раздела `Agents` из scaffold в production-ready lifecycle контур.
- Day1 intake (`#184`) зафиксировал проблему и границы MVP.
- Day2 vision (`#185`) зафиксировал mission/KPI и риск-рамку.
- Day3 PRD (`#187`) формализовал FR/AC/NFR-draft и создал handover issue в `run:arch` (`#189`).

## Эпики Sprint S6
- Day 1 (Intake): `docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`
- Day 2 (Vision baseline): GitHub issue `#185` (результаты vision использованы как вход в PRD).
- Day 3 (PRD):
  - `docs/delivery/epics/s6/epic-s6-day3-agents-prompts-prd.md`
  - `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md`
- Day 4 (Architecture issue): GitHub issue `#189` (`run:arch`).

## Планируемые epics (следующие stage)
- Day 4 (Architecture): C4/ADR/boundaries для agents/templates/audit domain.
- Day 5 (Design): API/data model/design package.
- Day 6 (Plan): Execution package и implementation issues.
- Day 7+ (Dev/QA/Release/Postdeploy/Ops + Doc-Audit): реализация, приемка и аудит трассируемости.

## Delivery-governance правила
- Каждый stage завершает работу созданием issue для следующего stage.
- Каждая следующая issue обязана содержать явную инструкцию создать issue после завершения текущего этапа.
- Для цепочки S6 зафиксирована последовательность continuity:
  - `#184 (intake) -> #185 (vision) -> #187 (prd) -> #189 (arch)`.
- До выхода в `run:dev` должны быть сформированы последовательные epics и связанные implementation issues.
