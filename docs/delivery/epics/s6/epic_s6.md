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
  request_id: "owner-2026-02-25-issue-184-intake"
---

# Epic Catalog: Sprint S6 (Agents configuration and prompt templates lifecycle)

## TL;DR
- Sprint S6 открывает полный stage-cycle для инициативы по реальному разделу `Agents` в staff UI и backend.
- Day1 intake фиксирует problem statement и acceptance baseline по текущему разрыву UI/Backend.
- Далее идут последовательные stage-epics без пропуска этапов, с обязательной генерацией follow-up issue после каждого stage.

## Эпики Sprint S6
- Day 1 (Intake): `docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`
- Day 2 (Vision issue): GitHub issue `#185` (ожидает завершения `run:vision`).
- Day 3 (PRD): `docs/delivery/epics/s6/epic-s6-day3-agents-prompts-prd.md` и `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md` (PR #190, ожидает merge).
- Day 4 (Architecture): `docs/delivery/epics/s6/epic-s6-day4-agents-prompts-arch.md`.

## Планируемые epics (будут добавлены на следующих stage)
- Day 5 (Design): API/data model/design package.
- Day 6 (Plan): Execution package и implementation issues.
- Day 7+ (Dev/QA/Release/Postdeploy/Ops + Doc-Audit): реализация, приемка и аудит трассируемости.

## Delivery-governance правила
- Каждый stage завершает работу созданием issue для следующего stage.
- Каждая следующая issue обязана содержать явную инструкцию создать issue после завершения текущего этапа.
- До выхода в `run:dev` должны быть сформированы последовательные epics и связанные implementation issues, как запросил Owner в Issue #184.
