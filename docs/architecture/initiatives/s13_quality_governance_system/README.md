---
doc_id: IDX-CK8S-ARCH-S13-0001
type: initiative-index
title: "Initiative Package: s13_quality_governance_system"
status: in-review
owner_role: SA
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [466, 469, 470, 471, 476, 484, 488, 494]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-484-arch"
---

# s13_quality_governance_system

## TL;DR
- Пакет объединяет Day4 architecture-артефакты Sprint S13 для `Quality Governance System`.
- Внутри зафиксированы ownership split для `control-plane` / `worker` / `api-gateway` / `web-console` / `agent-runner`, lifecycle `internal working draft -> semantic wave map -> published waves`, C4 overlays, ADR и alternatives по canonical change-governance aggregate.
- Следующий обязательный этап: Issue `#494` (`run:design`), где эти границы должны быть переведены в typed contracts, data model и rollout/migration policy без reopening policy semantics.

## Содержимое
- `docs/architecture/initiatives/s13_quality_governance_system/README.md`
- `docs/architecture/initiatives/s13_quality_governance_system/architecture.md`
- `docs/architecture/initiatives/s13_quality_governance_system/c4_context.md`
- `docs/architecture/initiatives/s13_quality_governance_system/c4_container.md`

## Связанные source-of-truth документы
- `docs/architecture/adr/ADR-0015-quality-governance-control-plane-owned-change-governance-aggregate.md`
- `docs/architecture/alternatives/ALT-0007-quality-governance-boundaries.md`
- `docs/delivery/epics/s13/epic-s13-day4-quality-governance-arch.md`
- `docs/delivery/epics/s13/epic-s13-day3-quality-governance-prd.md`
- `docs/delivery/epics/s13/prd-s13-day3-quality-governance-system.md`
- `docs/delivery/sprints/s13/sprint_s13_quality_governance_system.md`
- `docs/delivery/traceability/s13_quality_governance_system_history.md`

## Continuity after `run:arch`
- Документный контур `intake -> vision -> prd -> arch` согласован и доведён до review-ready architecture package.
- Issue `#494` остаётся единственным owner-managed handover в `run:design` и обязан сохранить этот пакет как baseline для typed contracts, projection model, migration policy и rollout sequencing.
- Sprint S14 (`#470`) остаётся downstream runtime/UI stream и может наследовать только typed surfaces из Day5/Day6, но не переоткрывать policy baseline Sprint S13.
