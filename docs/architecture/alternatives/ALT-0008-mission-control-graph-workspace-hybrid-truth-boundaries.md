---
doc_id: ALT-0008
type: alternatives
title: "Mission Control graph workspace — hybrid truth and ownership trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510, 516, 519]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-516-arch"
---

# Alternatives & Trade-offs: Mission Control graph workspace

## TL;DR
- Рассмотрели: GitHub/UI-defined graph / dedicated graph service now / control-plane-owned graph truth with worker-managed inventory foundation.
- Рекомендуем: `control-plane` как owner graph truth + `worker` для bounded provider foundation and reconcile execution.
- Почему: лучший баланс архитектурной консистентности, bounded scope Sprint S16, auditability и скорости handover в `run:design`.

## Контекст
- PRD Sprint S16 требует:
  - primary multi-root graph workspace;
  - issue `#480` как bounded provider foundation;
  - exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`;
  - node set `discussion/work_item/run/pull_request`;
  - typed metadata/watermarks, platform-canonical launch params и rule `PR + linked follow-up issue`.
- Нельзя нарушить:
  - thin-edge для `api-gateway` и `web-console`;
  - GitHub-native human review/merge semantics;
  - markdown-only scope на `run:arch`;
  - deferred boundary для voice/STT, dashboard orchestrator agent, отдельной `agent` taxonomy и full-history/archive.

## Вариант A: GitHub mirror + UI heuristics define the graph
- Описание:
  - считать graph workspace комбинацией provider mirror, live-fetch responses и client-side heuristics.
- Плюсы:
  - быстрый старт;
  - минимум backend decision surface на Day4.
- Минусы:
  - graph truth и continuity rule не имеют одного owner;
  - next-step surfaces становятся функцией UI logic.
- Риски:
  - split-brain между mirror, graph canvas и run lineage;
  - bounded history легко деградирует в непредсказуемый UX.
- Стоимость/сложность:
  - низкая на старте, высокая в стабилизации и аудите.

## Вариант B: Dedicated graph-workspace service уже сейчас
- Описание:
  - выделить новый runtime/service boundary для graph truth, projection serving и continuity orchestration.
- Плюсы:
  - чистый отдельный bounded context;
  - явный future scaling path.
- Минусы:
  - преждевременный service split до typed contracts и verified load signals;
  - дополнительный rollout и schema governance в середине doc-stage.
- Риски:
  - задержка `run:design` и `run:plan`;
  - новый coordination debt между `control-plane`, `worker` и сервисом.
- Стоимость/сложность:
  - высокая.

## Вариант C: Control-plane-owned graph truth + worker-managed inventory foundation (recommended)
- Описание:
  - `control-plane` владеет node kinds, relations, continuity gaps, metadata/watermarks и next-step policy;
  - `worker` исполняет bounded provider mirror sync, recent-closed-history backfill и enrichment/reconcile jobs;
  - `api-gateway`, `web-console`, `agent-runner` работают как typed adapters.
- Плюсы:
  - сохраняет текущие bounded contexts;
  - делает hybrid truth merge явным typed backend lifecycle;
  - удерживает bounded foundation `#480` и continuity rule в одном архитектурном контуре.
- Минусы:
  - увеличивает доменную ответственность `control-plane`;
  - требует аккуратной Day5 формализации watermarks и projection contracts.
- Риски:
  - future extraction всё равно может понадобиться при росте масштаба.
- Стоимость/сложность:
  - средняя.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Скорость Day4 handover | 5 | 1 | 4 |
| Архитектурная консистентность | 2 | 4 | 5 |
| Auditability | 2 | 4 | 5 |
| Риск scope drift | 1 | 2 | 5 |
| Готовность к `run:design` | 2 | 2 | 5 |
| Соответствие bounded foundation `#480` | 2 | 3 | 5 |

## Рекомендация
- Выбор: **Вариант C**.
- Обоснование:
  - сохраняет canonical owner для graph truth без premature new-service overhead;
  - удерживает provider mirror как bounded foundation, а не как substitute for graph semantics;
  - позволяет Day5 сфокусироваться на typed contracts/data model вместо переоткрытия ownership.
- Что теряем:
  - простоту варианта A и немедленную изоляцию варианта B.
- Что выигрываем:
  - воспроизводимый hybrid truth lifecycle, continuity completeness и более короткий путь к `run:design`.

## Нужен апрув от Owner
- [ ] Выбор варианта C.
- [ ] Разрешение компромисса: `control-plane` временно принимает ownership graph truth, а отдельный graph service откладывается до доказанных scale-сигналов.
