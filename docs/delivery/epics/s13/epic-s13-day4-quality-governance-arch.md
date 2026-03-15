---
doc_id: EPC-CK8S-S13-D4-QUALITY-GOVERNANCE
type: epic
title: "Epic S13 Day 4: Architecture для quality governance system и canonical governance ownership (Issues #484/#494)"
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

# Epic S13 Day 4: Architecture для quality governance system и canonical governance ownership (Issues #484/#494)

## TL;DR
- Подготовлен architecture package Sprint S13 для `Quality Governance System`: architecture decomposition, C4 overlays, ADR-0015 и alternatives по canonical governance aggregate, publication discipline и ownership split.
- Зафиксирован ownership split для risk/evidence/verification/waiver/publication semantics, asynchronous reconciliation, thin visibility surfaces и boundary `Sprint S13 -> Sprint S14`.
- Подготовлен handover в `run:design` без premature transport/schema lock-in.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#469` (`docs/delivery/epics/s13/epic-s13-day1-quality-governance-intake.md`).
- Vision baseline: `#471` (`docs/delivery/epics/s13/epic-s13-day2-quality-governance-vision.md`).
- PRD baseline: `#476` (`docs/delivery/epics/s13/epic-s13-day3-quality-governance-prd.md`, `docs/delivery/epics/s13/prd-s13-day3-quality-governance-system.md`).
- Текущий этап: `run:arch` в Issue `#484`.
- Scope этапа: только markdown-изменения.

## Architecture package
- `docs/architecture/initiatives/s13_quality_governance_system/README.md`
- `docs/architecture/initiatives/s13_quality_governance_system/architecture.md`
- `docs/architecture/initiatives/s13_quality_governance_system/c4_context.md`
- `docs/architecture/initiatives/s13_quality_governance_system/c4_container.md`
- `docs/architecture/adr/ADR-0015-quality-governance-control-plane-owned-change-governance-aggregate.md`
- `docs/architecture/alternatives/ALT-0007-quality-governance-boundaries.md`
- `docs/delivery/traceability/s13_quality_governance_system_history.md`

## Ключевые решения Stage
- `control-plane` остаётся единственным owner canonical change-governance aggregate, publication gate, risk/evidence/verification/waiver semantics и typed decision surface.
- `worker` закреплён за asynchronous reconciliation: freshness sweeps, stale-gate escalation, postdeploy feedback rollups и late reclassification; `agent-runner` не может быть owner policy semantics и остаётся signal emitter.
- `api-gateway` и `web-console` остаются thin transport/visibility surfaces и не вычисляют risk tier, evidence completeness, waiver rules или publication admissibility самостоятельно.
- Day4 закрепляет publication discipline `internal working draft -> semantic wave map -> published waves` и downstream rule: Sprint S14 (`#470`) наследует typed surfaces, но не переоткрывает policy baseline.

## Context7 и внешний baseline
- Проверен актуальный non-interactive GitHub CLI flow через Context7:
  - `/websites/cli_github_manual`.
- Локально проверены `gh issue create --help`, `gh pr create --help`, `gh pr edit --help` для non-interactive issue/PR automation.
- Новые внешние зависимости на этапе `run:arch` не требуются.

## Acceptance Criteria (Issue #484)
- [x] Подготовлен architecture package с service boundaries, ownership, C4 overlays, ADR и alternatives для `Quality Governance System`.
- [x] Для core flows определены owner-сервисы и границы ответственности: risk classification, evidence/verification evaluation, publication discipline, waiver/residual-risk decisions, asynchronous feedback reconciliation и visibility surfaces.
- [x] Зафиксированы architecture-level trade-offs по ownership и publication path без premature transport/storage lock-in.
- [x] Обновлены delivery/traceability документы и package indexes.
- [x] Подготовлена follow-up issue `#494` для stage `run:design` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S13-D4-01` Architecture completeness | Есть package `architecture + C4 + ADR + alternatives` | passed |
| `QG-S13-D4-02` Boundary integrity | Ownership за `control-plane` / `worker` / thin surfaces выражен явно | passed |
| `QG-S13-D4-03` Publication discipline | Путь `working draft -> semantic waves -> published waves` закреплён как domain lifecycle | passed |
| `QG-S13-D4-04` Proportionality integrity | Low-risk path и large/small change rules сохранены без policy drift | passed |
| `QG-S13-D4-05` Stage continuity | Подготовлена issue `#494` на `run:design` без trigger-лейбла | passed |

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#494`.
- Trigger-лейбл на новую issue ставит Owner после review architecture package.
- На design-stage обязательно:
  - выпустить `design_doc`, `api_contract`, `data_model`, `migrations_policy`;
  - определить typed ingress/projection/decision surfaces для change package, semantic wave map, evidence/verification/waiver states и postdeploy feedback;
  - зафиксировать rollout/backfill/rollback notes и transport/data boundaries без reopening policy baseline.
