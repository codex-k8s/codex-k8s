---
doc_id: EPC-CK8S-S13-D3-QUALITY-GOVERNANCE
type: epic
title: "Epic S13 Day 3: PRD для quality governance system в agent-scale delivery (Issues #476/#484)"
status: in-review
owner_role: PM
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [466, 469, 470, 471, 476, 484]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-476-prd"
---

# Epic S13 Day 3: PRD для quality governance system в agent-scale delivery (Issues #476/#484)

## TL;DR
- Подготовлен PRD-пакет Sprint S13 для `Quality Governance System`: `epic-s13-day3-quality-governance-prd.md` и `prd-s13-day3-quality-governance-system.md`.
- Зафиксированы user stories, FR/AC/NFR, edge cases, expected evidence и wave priorities для explicit risk tiering, evidence completeness, verification minimum, review/waiver discipline, proportional stage-gates, governance-gap feedback loop и publication policy `internal working draft -> semantic wave map -> published waves`.
- Принято продуктовое решение: Sprint S13 остаётся source-of-truth для change-governance contract, а downstream runtime/UI stream Sprint S14 (`#470`) наследует этот baseline и не переоткрывает его implementation-first.
- Создана follow-up issue `#484` для stage `run:arch` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#469` (`docs/delivery/epics/s13/epic-s13-day1-quality-governance-intake.md`).
- Vision baseline: `#471` (`docs/delivery/epics/s13/epic-s13-day2-quality-governance-vision.md`).
- Текущий этап: `run:prd` в Issue `#476`.
- Следующий этап: `run:arch` в Issue `#484`.

## Scope
### In scope
- Формализация user stories, FR/AC/NFR и edge cases для `Quality Governance System`.
- Приоритизация волн `core governance contract -> decision/waiver discipline -> deferred runtime/UI automation`.
- Фиксация product guardrails для explicit risk tiering, mandatory evidence package, verification minimum, residual-risk framing, proportional low-risk path, hidden `internal working draft`, semantic waves и high/critical no-silent-waiver policy.
- Явный handover в `run:arch` с перечнем продуктовых решений, которые нельзя потерять.
- Синхронизация traceability (`issue_map`, `delivery_plan`, sprint/epic docs, history package).

### Out of scope
- Кодовая реализация, storage/schema decisions и runtime/UI mechanics до `run:arch` / `run:design`.
- Выбор конкретных rollout, CI/CD, observability, quality cockpit или release-safety implementation.
- Попытка превратить Sprint S13 в general process bureaucracy redesign без привязки к measurable risk/evidence contract.
- Пересмотр или ослабление границы `Sprint S13 governance baseline -> Sprint S14 runtime/UI stream`.

## PRD package
- `docs/delivery/epics/s13/epic-s13-day3-quality-governance-prd.md`
- `docs/delivery/epics/s13/prd-s13-day3-quality-governance-system.md`
- `docs/delivery/traceability/s13_quality_governance_system_history.md`

## Wave priorities

| Wave | Приоритет | Scope | Exit signal |
|---|---|---|---|
| Wave 1 | `P0` | Explicit risk tier, mandatory evidence package, verification minimum, tier-aware completeness rules и publication rule `working draft -> semantic waves` | Каждый change package получает явный tier и не публикует raw draft в owner review |
| Wave 2 | `P0` | Review/waiver discipline, residual-risk framing, proportional stage-gates, role-specific decision surfaces и критерии допустимости больших/смешанных PR | Owner/reviewer, delivery roles и operator видят, что именно требуется для go/no-go и какие bundle допустимы |
| Wave 3 | `P1` (deferred) | Runtime/UI automation, quality cockpit, service-specific tuning и advanced policy automation | Stream попадает в roadmap только после подтверждения architecture/design package без reopening policy baseline |

## Acceptance criteria (Issue #476)
- [x] Подготовлен PRD-артефакт `Quality Governance System` и синхронизирован в traceability-документах.
- [x] Для core governance flows зафиксированы user stories, FR/AC/NFR, edge cases и expected evidence.
- [x] Wave priorities сформулированы без смешения core governance baseline и downstream runtime/UI automation stream Sprint S14.
- [x] Сохранены неподвижные ограничения инициативы: explicit risk tier, separate constructs `evidence completeness / verification minimum / review-waiver discipline`, proportional low-risk path, hidden `internal working draft`, semantic-wave publication policy и запрет silent waivers для `high/critical`.
- [x] Создана follow-up issue `#484` для stage `run:arch` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S13-D3-01 PRD completeness | User stories, FR/AC/NFR, edge cases и expected evidence покрывают scope Day3 | passed |
| QG-S13-D3-02 Guardrails preserved | Explicit risk tier, evidence/verification/waiver split, proportional governance и S13 -> S14 boundary сохранены | passed |
| QG-S13-D3-03 Deferred scope discipline | Runtime/UI automation и quality cockpit не смешаны с core product contract | passed |
| QG-S13-D3-04 Stage continuity | Создана issue `#484` для `run:arch` без trigger-лейбла | passed |
| QG-S13-D3-05 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#484`.
- Trigger-лейбл `run:arch` на issue `#484` ставит Owner после review PRD-пакета.
- Обязательные выходы архитектурного этапа:
  - service boundaries и ownership matrix для canonical change-governance aggregate, evidence-state lifecycle, waiver audit path и operator visibility;
  - alternatives/ADR по ownership `control-plane` / `worker` / `api-gateway` / `web-console` / `agent-runner` без потери product contract;
  - фиксация, как сохраняются proportional low-risk path, explicit high/critical discipline, publication path `working draft -> semantic waves` и boundary `Sprint S13 -> Sprint S14`;
  - отдельная issue на `run:design` без trigger-лейбла после завершения `run:arch`.

## Открытые риски и допущения
| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | `RSK-476-01` | Инициатива может расползтись в UI/process redesign и потерять core contract вокруг risk/evidence decisions | open |
| risk | `RSK-476-02` | Governance model окажется либо слишком жёсткой для `low`, либо слишком мягкой для `high/critical` | open |
| risk | `RSK-476-03` | Evidence completeness превратится в checkbox-theater без явного ownership и lifecycle на `run:arch` | open |
| assumption | `ASM-476-01` | Existing issue/PR/traceability surfaces достаточно зрелы для первой версии governance evidence и decision framing | accepted |
| assumption | `ASM-476-02` | Основная ценность достигается уже через explicit tiering, evidence contract и proportional gates без immediate runtime/UI automation | accepted |
