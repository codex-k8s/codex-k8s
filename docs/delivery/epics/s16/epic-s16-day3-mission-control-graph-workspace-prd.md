---
doc_id: EPC-CK8S-S16-D3-MISSION-CONTROL-GRAPH
type: epic
title: "Epic S16 Day 3: PRD для Mission Control graph workspace и continuity control plane (Issues #510/#516)"
status: superseded
owner_role: PM
created_at: 2026-03-16
updated_at: 2026-03-25
related_issues: [480, 490, 492, 496, 510, 516, 561, 562, 563]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-510-prd"
---

# Epic S16 Day 3: PRD для Mission Control graph workspace и continuity control plane (Issues #510/#516)

## TL;DR
- 2026-03-25 issue `#561` перевела этот PRD-пакет в historical superseded state.
- Зафиксированный здесь product contract больше не является текущим baseline для Mission Control и не должен использоваться для новых sprint issues.
- Актуальный baseline после rethink: frontend-first UX в `#562`, node taxonomy `Issue/PR/Run`, repo-seed prompts без DB prompt editor и backend rebuild в `#563`.
- Следующие разделы файла сохранены только как historical evidence отклонённого PRD.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#492` (`docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md`).
- Vision baseline: `#496` (`docs/delivery/epics/s16/epic-s16-day2-mission-control-graph-workspace-vision.md`).
- Текущий этап: `run:prd` в Issue `#510`.
- Следующий этап: `run:arch` в Issue `#516`.

## Scope
### In scope
- Формализация user stories, FR/AC/NFR, scenario matrix и expected evidence для Mission Control graph workspace и continuity control plane.
- Приоритизация волн `core graph workspace -> usability hardening -> deferred assistant/voice contours`.
- Фиксация product guardrails для exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity, closed Wave 1 node set `discussion/work_item/run/pull_request`, typed metadata/watermarks, platform-canonical launch params, platform-safe inline actions и continuity rule `PR + follow-up issue`.
- Явный handover в `run:arch` с перечнем продуктовых решений, которые нельзя потерять.
- Синхронизация traceability (`issue_map`, `delivery_plan`, sprint/epic docs, sprint history).

### Out of scope
- Кодовая реализация, выбор библиотек, storage/schema lock-in и transport/data details до `run:arch` / `run:design`.
- Возврат к board/list-first framing или GitHub-first single-root workspace как core product contract.
- Попытка сделать voice/STT, dashboard orchestrator agent, отдельную `agent` node taxonomy, full-history/archive или richer provider enrichment blocking scope для core Wave 1.
- Перенос human review, merge decision и provider-native collaboration из GitHub UI в dashboard.

## PRD package
- `docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md`
- `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`
- `docs/delivery/traceability/s16_mission_control_graph_workspace_history.md`

## Wave priorities

| Wave | Приоритет | Scope | Exit signal |
|---|---|---|---|
| Wave 1 | `P0` | Fullscreen graph workspace shell, filtered multi-root continuity graph, inventory-backed foundation, typed metadata/watermarks, platform-canonical launch params, platform-safe next-step actions, continuity rule `PR + follow-up issue` | Owner и delivery operator ведут 2-3 инициативы в одном workspace, видят lineage и понимают следующий безопасный шаг без ручного GitHub label hunting |
| Wave 2 | `P1` | Density/usability hardening, clearer continuity inspection surfaces, richer relation hints и non-blocking provider enrichments внутри уже зафиксированной truth model | Workspace остаётся читаемым при росте active set и не теряет trust к coverage/freshness |
| Wave 3 | `P1` (deferred) | Voice/STT, dashboard orchestrator agent, отдельная `agent` node taxonomy, full-history/archive, richer provider enrichment beyond agreed baseline | Stream двигается только после owner-approved architecture/design без reopening core Wave 1 contract |

## Acceptance criteria (Issue #510)
- [x] Подготовлен PRD-артефакт Mission Control graph workspace и синхронизирован в traceability-документах.
- [x] Для core flows зафиксированы user stories, FR/AC/NFR, scenario matrix, edge cases и expected evidence.
- [x] Явно сохранены locked baselines: issue `#480`, exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity, Wave 1 nodes `discussion/work_item/run/pull_request`, typed metadata/watermarks, platform-canonical launch params и continuity rule `PR + follow-up issue`.
- [x] Wave priorities сформулированы без смешения core Wave 1 и deferred voice/orchestrator/agent/full-history contours.
- [x] Создана follow-up issue `#516` для stage `run:arch` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S16-D3-01 PRD completeness | User stories, FR/AC/NFR, scenario matrix и expected evidence покрывают scope Day3 | passed |
| QG-S16-D3-02 Locked baseline preserved | `#480`, exact filters/nodes, typed metadata/watermarks, launch params и continuity rule сохранены без reopening | passed |
| QG-S16-D3-03 Deferred-scope discipline | Voice/STT, orchestrator path, отдельные `agent` nodes, full-history/archive и richer provider enrichment не смешаны с core Wave 1 | passed |
| QG-S16-D3-04 Stage continuity | Создана issue `#516` для `run:arch` без trigger-лейбла и с continuity-требованием `arch -> design -> plan -> dev` | passed |
| QG-S16-D3-05 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#516`.
- Trigger-лейбл `run:arch` на issue `#516` ставит Owner после review PRD-пакета.
- Обязательные выходы архитектурного этапа:
  - service boundaries и ownership matrix для `control-plane`, `worker`, `api-gateway`, `web-console` вокруг graph truth, continuity state, provider mirror freshness и typed launch/metadata surfaces;
  - alternatives/ADR по hybrid truth merge, persisted projections/relations, provider-safe inline actions и continuity graph без потери product contract;
  - фиксация, как сохраняются exact Wave 1 filters/nodes, secondary/dimmed semantics, coverage contract `all open Issues/PR + bounded recent closed history`, typed watermarks и boundary `core Wave 1 -> deferred contours`;
  - отдельная issue на `run:design` без trigger-лейбла после завершения `run:arch` с явным continuity-требованием продолжить цепочку `design -> plan -> dev`.

## Открытые риски и допущения
| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | `RSK-510-01` | Инициатива может расползтись в смесь graph UX, provider mirror, voice/orchestrator path и новой node taxonomy | open |
| risk | `RSK-510-02` | Multi-root workspace без чёткой truth/ownership model создаст duplicate nodes, missing lineage или недоверие к next-step actions | open |
| risk | `RSK-510-03` | Continuity rule `PR + follow-up issue` может деградировать в необязательную рекомендацию без явной product/architecture фиксации | open |
| assumption | `ASM-510-01` | Existing bounded contexts `control-plane` / `worker` / `api-gateway` / `web-console` можно эволюционно расширить до graph workspace без нового deployable сервиса на этой стадии | accepted |
| assumption | `ASM-510-02` | Основная пользовательская ценность достигается уже на core Wave 1 без voice/STT, dashboard orchestrator agent и отдельной `agent` taxonomy | accepted |
