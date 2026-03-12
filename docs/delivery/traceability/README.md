---
doc_id: IDX-CK8S-DELIVERY-TRACE-0001
type: traceability-index
title: "Delivery Traceability History Index"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [325, 327]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-delivery-traceability-index"
---

# Delivery Traceability History

## TL;DR
- `docs/delivery/issue_map.md` остаётся master-index для текущей карты `issue -> docs -> status`.
- `docs/delivery/requirements_traceability.md` остаётся стабильной FR/NFR-матрицей текущего состояния.
- `docs/delivery/traceability/*.md` хранит historical evidence и delta по спринтам, чтобы root-реестры не смешивали актуальное состояние и execution-history.

## Структура

| Файл | Sprint / scope | Что хранит |
|---|---|---|
| `docs/delivery/traceability/s5_stage_entry_and_label_ux_history.md` | Sprint S5 | История plan/dev-evidence по launch profiles, next-step actions и reviewer pre-review policy |
| `docs/delivery/traceability/s6_agents_prompt_management_history.md` | Sprint S6 | История полного stage-контура `intake -> dev -> release -> postdeploy -> ops` для lifecycle agents/prompts |
| `docs/delivery/traceability/s7_mvp_readiness_gap_closure_history.md` | Sprint S7 | История stage/development evidence по MVP readiness execution streams |
| `docs/delivery/traceability/s8_go_refactoring_parallelization_history.md` | Sprint S8 | История go-refactor/onboarding/doc-governance потоков, включая doc-audit decomposition issue `#327` |

## Правила обновления
- В `docs/delivery/issue_map.md` обновляется только текущая каноническая карта связей, bundle refs и статус issue/PR.
- В `docs/delivery/requirements_traceability.md` обновляется только актуальное покрытие FR/NFR и правило поддержания матрицы.
- В sprint history file добавляется delta, если stage/issue оставляет historical evidence, continuity-решение, quality-gate outcome или remediation-note, которые важно сохранить, но не нужно держать в root-реестрах.
- History package не должен переопределять текущий source of truth и не должен дублировать всю root-матрицу целиком.

## Правило декомпозиции
- Базовая единица хранения historical evidence: один markdown-файл на один sprint.
- Для нового спринта создаётся отдельный `docs/delivery/traceability/s<номер>_*.md`, когда в нём появляется более одного issue-specific update или когда root-реестры начинают смешивать stable state и execution-history.
- Если issue тянется через несколько спринтов, evidence добавляется в файл того спринта, где сформирован конкретный delta/result; cross-sprint continuity остаётся в `issue_map.md`.

## Migration Note
- В Issue `#327` исторические секции `## Актуализация по Issue ...` вынесены из `docs/delivery/requirements_traceability.md` в sprint-specific history packages.
- Root delivery traceability после этого разделена по уровням абстракции: current state в корневых реестрах, historical evidence в `docs/delivery/traceability/`.
