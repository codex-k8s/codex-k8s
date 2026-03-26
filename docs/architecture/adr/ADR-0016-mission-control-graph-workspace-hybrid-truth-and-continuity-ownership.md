---
doc_id: ADR-0016
type: adr
title: "Mission Control graph workspace: control-plane-owned hybrid graph truth with worker-managed inventory foundation"
status: superseded
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-25
related_issues: [480, 490, 492, 496, 510, 516, 519, 561, 562, 563]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-516-arch"
---

# ADR-0016: Mission Control graph workspace — control-plane-owned hybrid graph truth with worker-managed inventory foundation

## TL;DR
- 2026-03-25 issue `#561` перевела ADR-0016 в historical superseded state.
- Документ сохраняется как evidence отклонённого Sprint S16 baseline и не является текущим architecture candidate для следующего backend sprint `#563`.
- Новый baseline после rethink требует сначала утвердить frontend-first UX и workflow editor contour в issue `#562`, а backend truth/model переопределять только после owner approval.
- Контекст: Sprint S16 PRD требует primary multi-root graph workspace, bounded provider foundation `#480`, exact Wave 1 filters/nodes, typed metadata/watermarks, platform-canonical launch params и continuity rule `PR + linked follow-up issue`.
- Решение: выбираем `control-plane` как единственного owner canonical graph truth и continuity state; `worker` исполняет bounded provider inventory sync, recent-closed-history backfill и enrichment/reconcile jobs, а `api-gateway` / `web-console` / `agent-runner` остаются adapters над typed surfaces.
- Последствия: появляется один domain owner для hybrid truth merge и next-step policy, но design-stage обязан отдельно зафиксировать transport/data contracts, rollout/backfill notes и watermark taxonomy.

## Status after rethink
- ADR-0016 больше не описывает текущий source of truth для Mission Control.
- Superseded assumptions:
  - lane/column-oriented Sprint S16 graph workspace как целевой UX baseline;
  - taxonomy `discussion/work_item/run/pull_request` как Wave 1 node model;
  - backend-first sequencing вокруг execution handover `#542..#547`.
- Historical value документа сохраняется только как evidence того, какой ownership split и hybrid-truth lifecycle были предложены до owner reset.

## Контекст
- Проблема:
  - без закреплённого owner hybrid truth Sprint S16 расползётся между GitHub mirror, UI heuristics, background jobs и run-local context;
  - rule `PR + linked follow-up issue` требует persisted domain semantics, иначе continuity снова станет narrative convention;
  - попытка сделать GitHub mirror или отдельный graph service primary owner раньше `run:design` размоет bounded contexts и затянет handover.
- Ограничения:
  - `run:arch` остаётся markdown-only;
  - `api-gateway` и `web-console` должны оставаться thin adapters;
  - issue `#480` остаётся bounded provider foundation, а не full-history warehouse;
  - voice/STT, dashboard orchestrator agent, отдельная `agent` node taxonomy, full-history/archive и richer provider enrichment остаются deferred.
- Связанные требования:
  - `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`
  - `FR-002`, `FR-003`, `FR-004`, `FR-009`, `FR-012`, `FR-026`, `FR-027`
  - `NFR-010`, `NFR-018`
- Что ломается без решения:
  - разные части системы будут по-разному определять node kind, graph completeness и allowed next step;
  - Day5 придётся заново спорить об owner-сервисе вместо typed contracts;
  - bounded inventory foundation `#480` либо превратится в full-history warehouse, либо будет недооценена как необязательная.

## Decision Drivers
- Единый domain owner для graph truth, continuity lineage и next-step semantics.
- Сохранение GitHub как canonical provider source для issue/pr/comment/review state.
- Bounded inventory foundation без превращения в отдельный продукт.
- Thin-edge consistency для `api-gateway` и `web-console`.
- Минимизация premature service split до появления доказанных scale signals.

## Рассмотренные варианты

### Вариант A: GitHub mirror and UI heuristics define the graph
- Плюсы:
  - минимальный стартовый cost;
  - можно быстро собрать визуализацию поверх provider data.
- Минусы:
  - graph semantics зависят от client heuristics и mirror completeness;
  - continuity rule `PR + linked follow-up issue` не становится canonical state.
- Риски:
  - split-brain между UI, mirror и run artifacts;
  - потеря доверия к next-step surfaces;
  - scope drift в сторону live-fetch-only dashboard.
- Стоимость внедрения:
  - низкая на старте, высокая по rework и аудиту.
- Эксплуатация:
  - трудно доказать, какая ветка реально complete, а какая имеет continuity gap.

### Вариант B: Новый dedicated graph-workspace service уже на Day4
- Плюсы:
  - отдельный bounded context;
  - потенциальный future scaling path.
- Минусы:
  - новый DB owner и rollout contour до фиксации typed contracts;
  - дополнительный integration burden в середине markdown-only stage.
- Риски:
  - задержка `run:design` и `run:plan`;
  - преждевременный lock-in в service topology.
- Стоимость внедрения:
  - высокая.
- Эксплуатация:
  - ещё один coordination path без подтверждённой runtime необходимости.

### Вариант C (выбран): `control-plane` owns graph truth, `worker` owns bounded inventory execution, thin surfaces consume typed projections
- Плюсы:
  - сохраняет текущие service boundaries платформы;
  - даёт один owner для node classification, continuity state и hybrid truth merge;
  - удерживает `worker`, `api-gateway`, `web-console` и `agent-runner` в роли adapters/executors.
- Минусы:
  - `control-plane` получает дополнительную доменную ответственность;
  - design-stage должен аккуратно разложить graph truth, mirror references и watermark taxonomy.
- Риски:
  - при росте throughput может потребоваться future extraction read-model/service layer.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - предсказуемая при условии typed contracts и audit-safe reconcile semantics.

## Решение
Мы выбираем: **вариант C — `control-plane` owns canonical graph truth and continuity state, `worker` executes bounded inventory foundation and reconcile, thin surfaces consume typed projections**.

## Обоснование (Rationale)
- Вариант C лучше всего соответствует текущей архитектуре `codex-k8s`:
  - `control-plane` уже владеет stage policy, run lifecycle, audit и launch semantics;
  - `worker` естественно подходит для mirror sync, recent-closed-history backfill и enrichment/reconcile execution, но не для canonical graph semantics;
  - `api-gateway` и `web-console` сохраняют thin-edge/presentation роли;
  - `agent-runner` остаётся source emitter, а не source-of-truth.
- Решение сохраняет product baseline Sprint S16 и не переоткрывает Sprint S9 implementation choices: active-set dashboard becomes subordinate to multi-root graph truth, а not vice versa.
- Bounded inventory foundation `#480` получает устойчивое место в архитектуре без превращения в отдельный warehouse stream.

## Последствия (Consequences)

### Позитивные
- Появляется единый owner graph truth, continuity lineage и next-step semantics.
- Hybrid truth merge можно проектировать как явный typed lifecycle `provider mirror -> graph truth -> workspace projection`.
- Continuity rule `PR + linked follow-up issue` становится проверяемым persisted construct, а не только traceability note.

### Негативные / компромиссы
- `control-plane` получает больше доменной логики вокруг graph workspace.
- Design-stage должен отдельно доказать, что watermarks, bounded history policy и list fallback не размоют core graph truth.

### Технический долг
- Что откладываем:
  - отдельный graph workspace service;
  - voice/STT, dashboard orchestrator agent и отдельную `agent` node taxonomy;
  - richer provider enrichment beyond agreed bounded mirror.
- Когда вернуться:
  - после `run:design` / `run:plan`, если появятся явные scale или ownership breakdown signals.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - `run:design` должен определить canonical graph data model, mirror references, watermarks, continuity gaps и backfill strategy.
- Наблюдаемость:
  - `run:design` должен зафиксировать event set для continuity gap detection, launch preview/launch, mirror freshness and graph watermark updates.

## План отката/замены
- Условия отката:
  - если на `run:design` выяснится, что graph truth не удерживается в `control-plane` без unacceptable coupling или throughput collapse.
- Как откатываем:
  - ADR переводится в `superseded`, а graph-truth/read-model path выделяется в отдельный service boundary с сохранением тех же Sprint S16 guardrails.

## Ссылки
- PRD:
  - `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`
- Architecture:
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0008-mission-control-graph-workspace-hybrid-truth-boundaries.md`
- Historical baseline:
  - `docs/architecture/adr/ADR-0011-mission-control-dashboard-active-set-projection-and-command-reconciliation.md`
  - `docs/architecture/initiatives/s9_mission_control_dashboard/architecture.md`
