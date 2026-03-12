---
doc_id: EPC-CK8S-S9-D4-MISSION-CONTROL
type: epic
title: "Epic S9 Day 4: Architecture для Mission Control Dashboard и console control plane (Issue #340)"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340, 351]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-340-arch"
---

# Epic S9 Day 4: Architecture для Mission Control Dashboard и console control plane (Issue #340)

## TL;DR
- Подготовлен architecture package Sprint S9 для Mission Control Dashboard: architecture decomposition, C4 overlays, ADR-0011 и alternatives по projection/realtime ownership.
- Зафиксирован ownership split для active-set projection, timeline/comments mirror, typed commands, provider sync/reconciliation, degraded realtime path и isolated voice candidate stream.
- Подготовлен handover в `run:design` без premature library lock-in.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#333` (`docs/delivery/epics/s9/epic-s9-day1-mission-control-dashboard-intake.md`).
- Vision baseline: `#335` (`docs/delivery/epics/s9/epic-s9-day2-mission-control-dashboard-vision.md`).
- PRD baseline: `#337` (`docs/delivery/epics/s9/epic-s9-day3-mission-control-dashboard-prd.md`, `docs/delivery/epics/s9/prd-s9-day3-mission-control-dashboard.md`).
- Текущий этап: `run:arch` в Issue `#340`.
- Scope этапа: только markdown-изменения.

## Architecture package
- `docs/architecture/initiatives/s9_mission_control_dashboard/README.md`
- `docs/architecture/initiatives/s9_mission_control_dashboard/architecture.md`
- `docs/architecture/initiatives/s9_mission_control_dashboard/c4_context.md`
- `docs/architecture/initiatives/s9_mission_control_dashboard/c4_container.md`
- `docs/architecture/adr/ADR-0011-mission-control-dashboard-active-set-projection-and-command-reconciliation.md`
- `docs/architecture/alternatives/ALT-0003-mission-control-dashboard-projection-and-realtime-trade-offs.md`
- `docs/delivery/traceability/s9_mission_control_dashboard_history.md`

## Ключевые решения Stage
- `control-plane` остаётся владельцем active-set projection, relation graph, timeline/comments mirror и command lifecycle.
- `worker` закреплён за provider sync, retries и reconciliation loops; `api-gateway` не принимает решений о dedupe/policy.
- Snapshot-first / delta-second realtime baseline подтверждён как обязательный UX contract; degraded state обязан поддерживать explicit refresh и list fallback.
- Voice intake изолирован как optional candidate stream и не может загрязнять core MVP page-load/contracts.

## Context7 верификация
- Проверен актуальный синтаксис GitHub CLI для follow-up issue и PR flow:
  - `/websites/cli_github_manual`.
- Проверен актуальный Mermaid C4 syntax для диаграмм пакета:
  - `/mermaid-js/mermaid`.
- Новые внешние зависимости в `run:arch` не требуются.

## Acceptance Criteria (Issue #340)
- [x] Подготовлен architecture package с service boundaries, ownership, C4 overlays, ADR и alternatives для Mission Control Dashboard.
- [x] Для core flows определены owner-сервисы и границы ответственности: active-set dashboard, discussion formalization, commands/reconciliation, timeline/comments mirror, degraded mode.
- [x] Зафиксированы architecture-level trade-offs по realtime transport, projections и voice candidate stream без premature library lock-in.
- [x] Обновлены delivery/traceability документы и package indexes.
- [x] Подготовлена follow-up issue `#351` для stage `run:design` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S9-D4-01` Architecture completeness | Есть package `architecture + C4 + ADR + alternatives` | passed |
| `QG-S9-D4-02` Boundary integrity | Thin-edge сохранён, ownership за `control-plane`/`worker` зафиксирован явно | passed |
| `QG-S9-D4-03` Realtime fallback discipline | Snapshot-first / delta-second / degraded fallback описаны как единый контракт | passed |
| `QG-S9-D4-04` Voice isolation | Voice stream не блокирует core MVP и не входит в primary contracts | passed |
| `QG-S9-D4-05` Stage continuity | Подготовлена issue на `run:design` без trigger-лейбла | passed |

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#351`.
- Trigger-лейбл на новую issue ставит Owner после review architecture package.
- На design-stage обязательно:
  - выпустить `design_doc`, `api_contract`, `data_model`, `migrations_policy`;
  - определить typed contracts для snapshot, entity details, commands, command status и realtime topics;
  - зафиксировать projection persistence strategy, migration order и rollback notes.
