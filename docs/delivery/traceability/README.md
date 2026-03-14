---
doc_id: IDX-CK8S-DELIVERY-TRACE-0001
type: traceability-index
title: "Delivery Traceability History Index"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-14
related_issues: [325, 327, 333, 335, 337, 340, 351, 360, 361, 363, 366, 369, 370, 371, 372, 373, 374, 375, 378, 383, 385, 387, 413, 416, 418, 444, 447, 448]
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
| `docs/delivery/traceability/s9_mission_control_dashboard_history.md` | Sprint S9 | История intake/vision/prd/arch/design/plan решений по Mission Control Dashboard, включая execution backlog `#369..#375` и split core vs conditional voice contour |
| `docs/delivery/traceability/s10_mcp_user_interactions_history.md` | Sprint S10 | История intake, vision, PRD и architecture baseline по built-in MCP user interactions, включая continuity issues `#385` (`run:arch`) и `#387` (`run:design`) |
| `docs/delivery/traceability/s11_telegram_user_interaction_adapter_history.md` | Sprint S11 | История intake и vision baseline по Telegram-адаптеру как первому внешнему channel path, включая closed superseded handover issue `#444`, active vision stage `#447` и continuity issue `#448` для `run:prd` |
| `docs/delivery/traceability/s12_github_api_rate_limit_resilience_history.md` | Sprint S12 | История intake/vision/PRD решений по GitHub API rate-limit resilience, включая continuity issues `#416` (`run:prd`) и `#418` (`run:arch`) |

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
