---
doc_id: EPC-CK8S-S9-D3-MISSION-CONTROL
type: epic
title: "Epic S9 Day 3: PRD для Mission Control Dashboard и console control plane (Issues #337/#340)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-337-prd"
---

# Epic S9 Day 3: PRD для Mission Control Dashboard и console control plane (Issues #337/#340)

## TL;DR
- Подготовлен PRD-пакет Sprint S9 для Mission Control Dashboard: `epic-s9-day3-mission-control-dashboard-prd.md` и `prd-s9-day3-mission-control-dashboard.md`.
- Зафиксированы user stories, FR/AC/NFR, edge cases, expected evidence и wave priorities для active-set dashboard, discussion formalization, provider-safe commands и reconciliation guardrails.
- Принято продуктовое решение: voice intake остаётся условной Wave 3 и не блокирует core MVP; GitHub-first MVP, external human review и active-set default остаются неподвижными ограничениями.
- Создана follow-up issue `#340` для stage `run:arch` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#333` (`docs/delivery/epics/s9/epic-s9-day1-mission-control-dashboard-intake.md`).
- Vision baseline: `#335` (`docs/delivery/epics/s9/epic-s9-day2-mission-control-dashboard-vision.md`).
- Текущий этап: `run:prd` в Issue `#337`.
- Следующий этап: `run:arch` в Issue `#340`.

## Scope
### In scope
- Формализация user stories, FR/AC/NFR и edge cases для Mission Control Dashboard.
- Приоритизация волн `pilot -> MVP release -> conditional candidate stream`.
- Фиксация product guardrails для active-set UX, discussion-first flow, provider-safe commands и reconciliation.
- Явный handover в `run:arch` с перечнем product decisions, которые нельзя потерять.
- Синхронизация traceability (`issue_map`, `delivery_plan`, sprint/epic docs, history package).

### Out of scope
- Кодовая реализация, выбор библиотек и transport/storage lock-in.
- Детальный UI layout spec и implementation-ready API/data model.
- Перенос human review и merge decision из provider UI в staff console.
- Возврат voice intake или GitLab parity в blocking scope core MVP.

## PRD package
- `docs/delivery/epics/s9/epic-s9-day3-mission-control-dashboard-prd.md`
- `docs/delivery/epics/s9/prd-s9-day3-mission-control-dashboard.md`
- `docs/delivery/traceability/s9_mission_control_dashboard_history.md`

## Wave priorities

| Wave | Приоритет | Scope | Exit signal |
|---|---|---|---|
| Wave 1 | `P0` | Active-set dashboard shell, typed entity/relation baseline, side panel, filters/search, list fallback, provider context visibility | Пользователь за 5-10 секунд понимает active set и может выбрать следующее действие из одного control-plane экрана |
| Wave 2 | `P0` | Discussion-first formalization, comments/timeline projection, command status, provider sync, webhook echo dedupe/correlation, degraded realtime fallback | Discussion -> formal task и console-initiated actions работают без дублей и без потери traceability |
| Wave 3 | `P1` (conditional) | Guided voice intake и AI-assisted draft structuring | Stream входит в roadmap только если PRD/architecture подтверждают ROI, policy fit и operational readiness |

## Acceptance criteria (Issue #337)
- [x] Подготовлен PRD-артефакт Mission Control Dashboard и синхронизирован в traceability-документах.
- [x] Для core flows зафиксированы user stories, FR/AC/NFR, edge cases и expected evidence.
- [x] Wave priorities сформулированы без смешения core MVP и conditional voice stream.
- [x] Сохранены неподвижные ограничения инициативы: GitHub-first MVP, human review во внешнем provider UI, active-set default, label/audit policy и webhook-driven orchestration.
- [x] Создана follow-up issue `#340` для stage `run:arch` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S9-D3-01 PRD completeness | User stories, FR/AC/NFR, edge cases и expected evidence покрывают scope Day3 | passed |
| QG-S9-D3-02 Wave discipline | Core MVP и conditional voice stream разделены по приоритетам и exit signals | passed |
| QG-S9-D3-03 Guardrails preserved | GitHub-first MVP, external review, active-set default и policy-safe commands сохранены | passed |
| QG-S9-D3-04 Stage continuity | Создана issue `#340` для `run:arch` без trigger-лейбла | passed |
| QG-S9-D3-05 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#340`.
- Trigger-лейбл `run:arch` на issue `#340` ставит Owner после review PRD-пакета.
- Обязательные выходы архитектурного этапа:
  - service boundaries и ownership для dashboard read model, typed commands, provider sync, reconciliation и comments/timeline projection;
  - alternatives/ADR по realtime transport, active-set projection и degraded-mode path без потери product contract;
  - фиксация, как сохраняются active-set default, provider-safe commands и external human review;
  - отдельная issue на `run:design` без trigger-лейбла после завершения `run:arch`.

## Открытые риски и допущения
| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | RSK-337-01 | Инициатива может расползтись в общий redesign staff console вместо control-plane UX | open |
| risk | RSK-337-02 | High-volume active set потребует более жёстких ограничений по density/grouping, чем ожидается сейчас | open |
| risk | RSK-337-03 | Command/reconciliation path может стать критическим bottleneck для delivery, если ownership останется неявным | open |
| assumption | ASM-337-01 | GitHub-first provider model достаточно покрывает первую волну Mission Control Dashboard | accepted |
| assumption | ASM-337-02 | Существующий realtime baseline можно расширять, а не проектировать заново | accepted |

