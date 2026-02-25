---
doc_id: EPC-CK8S-S5-D2
type: epic
title: "Epic S5 Day 2: Single-epic execution for launch profiles and deterministic next-step actions (Issues #170/#171)"
status: approved
owner_role: EM
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [170, 171, 154, 155]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-170-single-epic"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic S5 Day 2: Single-epic execution for launch profiles and deterministic next-step actions (Issues #170/#171)

## TL;DR
- Day2 фиксирует единый контур реализации FR-053/FR-054: один эпик и одна implementation issue.
- План исполнения переведён в delivery-ready формат с quality-gates, DoD и handover по ролям.
- GitHub implementation issue создана: #171.

## Контекст и цель
- Родительский planning-контур: Issue #170 (`run:plan`).
- Execution issue: #171 (single-epic implementation scope).
- Источники требований и архитектурных ограничений:
  - `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`
  - `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`
  - `docs/architecture/adr/ADR-0008-profile-driven-stage-launch-and-next-step-contract.md`
  - `docs/product/requirements_machine_driven.md` (FR-053, FR-054)

Цель Day2: подготовить детерминированный и проверяемый execution-пакет для входа в `run:dev` без изменения согласованных архитектурных границ.

## Scope
### In scope
- I1 (`P0`): deterministic profile resolver (`quick-fix|feature|new-service`) + escalation rules.
- I2 (`P0`): typed next-step action-card contract (`launch_profile`, `stage_path`, `primary_action`, `fallback_action`, `guardrail_note`).
- I3 (`P0`): fallback pre-check + transition path, ambiguity hard-stop (`need:input`).
- I4 (`P1`): review-gate sync для пары `Issue + PR` (`state:in-review`).
- I5 (`P1`): traceability sync после переходов (`issue_map`, `requirements_traceability`, sprint/epic docs).

### Out of scope
- Изменение базовой taxonomy labels вне `run:*|state:*|need:*`.
- Пересмотр RBAC/namespace retention policy.
- Изменения multi-repo контура Sprint S4.

## Quality-gates
| Gate | Что проверяем | Критерий выхода |
|---|---|---|
| QG-D2-01 Planning | Single-epic scope зафиксирован между #170 и #171 | Один delivery-эпик и один implementation issue, без параллельных dev-epics |
| QG-D2-02 Contract | Next-step contract соответствует ADR-0008 и API-contract | Все обязательные поля action-card определены и трассируемы |
| QG-D2-03 Governance | Fallback path policy-safe | Есть pre-check, ambiguity-stop и запрет best-guess |
| QG-D2-04 Architecture | Service boundaries не нарушены | Resolver/escalation остаётся в `control-plane`, edge/UI thin |
| QG-D2-05 Traceability | FR-053/FR-054 и issue-map синхронизированы | Документы Sprint S5, issue_map и RTM обновлены |
| QG-D2-06 Readiness | Пакет готов к `run:dev` | Owner-ready handover для `dev/qa/sre/km` сформирован |

## Definition of Done (Day2 planning package)
- [x] Создан отдельный Day2 epic-документ с single-epic execution-моделью.
- [x] Создана GitHub implementation issue #171 и связана с #170.
- [x] Обновлены Sprint S5 plan и Epic catalog (`epic_s5.md`).
- [x] Обновлены `issue_map` и `requirements_traceability` для FR-053/FR-054.
- [x] Зафиксированы блокеры, риски и owner decisions для handover в `run:dev`.

## Acceptance baseline для `run:dev`
- AC-01..AC-06 из Day1 epic остаются обязательными.
- Дополнительно для single-epic исполнения:
  - AC-D2-01: все инкременты I1..I5 выполняются в рамках одной implementation issue (#171);
  - AC-D2-02: приёмка не допускает partial-handover без traceability sync;
  - AC-D2-03: regression evidence по AC-01..AC-06 привязан к PR и отражён в review-gate.

## Blockers, риски и owner decisions
| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | BLK-170-01 | Требовалось подтвердить формат single-epic execution (один эпик, одна implementation issue) | closed |
| blocker | BLK-170-02 | Требовалось закрепить обязательный pre-check перед fallback transition | closed |
| risk | RSK-170-01 | Рост объёма работ в одной issue может замедлить review cycle | monitoring |
| risk | RSK-170-02 | Drift между fallback-командами и runtime policy | monitoring |
| owner-decision | OD-170-01 | Реализация идёт одним эпиком и одной implementation issue (#171) | approved |
| owner-decision | OD-170-02 | Все связанные документы Day2 фиксируются со статусом согласовано | approved |

## Handover по ролям
- `dev`: выполнить I1..I5 в рамках #171 и сформировать PR с evidence по AC-01..AC-06.
- `qa`: подготовить и подтвердить regression-пакет по AC-01..AC-06 + AC-D2-01..AC-D2-03.
- `sre`: проверить audit path и исключить policy bypass для fallback transitions.
- `km`: удерживать синхронизацию traceability после каждого merge/revise.

## Связанные артефакты
- `docs/delivery/sprints/s5/sprint_s5_stage_entry_and_label_ux.md`
- `docs/delivery/epics/s5/epic_s5.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/requirements_traceability.md`
- GitHub issue #171
