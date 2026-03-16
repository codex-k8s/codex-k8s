---
doc_id: EPC-CK8S-S16-D6-MISSION-CONTROL-GRAPH
type: epic
title: "Epic S16 Day 6: Plan для Mission Control graph workspace, execution waves и rollout gates (Issue #537)"
status: in-review
owner_role: EM
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510, 516, 519, 537, 542, 543, 544, 545, 546, 547]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-537-plan-epic"
---

# Epic S16 Day 6: Plan для Mission Control graph workspace, execution waves и rollout gates (Issue #537)

## TL;DR
- Подготовлен execution package Sprint S16 для перехода в `run:dev` по Mission Control graph workspace.
- Созданы owner-managed handover issues `#542..#547` для schema/backfill foundation, `control-plane` graph truth, `worker` reconcile, typed transport, `web-console` graph workspace и финального readiness gate.
- Зафиксированы sequencing-waves, quality-gates, DoR/DoD и owner decisions для rollout order `migrations -> control-plane -> worker -> api-gateway -> web-console -> readiness gate`.
- Сохранены design guardrails Sprint S16: issue `#480` остаётся foundation layer, exact Wave 1 filters `open_only` и `assigned_to_me_or_unassigned` не меняются, active-state presets и secondary/dimmed semantics работают только для graph integrity, nodes ограничены `discussion/work_item/run/pull_request`, launch preview остаётся read-only поверх existing command ledger, новый deployable сервис запрещён.

## Контекст
- Stage continuity: `#492 -> #496 -> #510 -> #516 -> #519 -> #537`.
- Входной baseline: design package Day5
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/design_doc.md`
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/api_contract.md`
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/data_model.md`
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/migrations_policy.md`
- Scope текущего stage: только markdown-изменения и handover backlog.
- Правило continuity: trigger-лейблы `run:dev` на implementation issues ставит только Owner и только по wave-sequencing.

## Execution package (S16-E01..S16-E06)

| Stream | Implementation issue | Wave | Priority | Краткий scope |
|---|---:|---|---|---|
| `S16-E01` | #542 | Wave 1 | P0 | Additive schema, graph/continuity foundation и bounded shadow backfill под coverage contract issue `#480` |
| `S16-E02` | #543 | Wave 2 | P0 | `control-plane` graph truth, workspace projections, continuity-gap lifecycle и read-only launch preview semantics |
| `S16-E03` | #544 | Wave 3 | P0 | `worker` reconcile, inventory freshness, warmup/backfill execution и parity signals без ownership graph semantics |
| `S16-E04` | #545 | Wave 4 | P0 | Contract-first `api-gateway` transport surfaces, typed DTO/casters и launch preview exposure |
| `S16-E05` | #546 | Wave 5 | P0 | `web-console` graph-first workspace, continuity UX, linked artifact visibility и preview presentation |
| `S16-E06` | #547 | Wave 6 | P0 | Observability, rollout/rollback discipline и readiness evidence gate перед `run:qa` |

## Sequencing constraints
- Wave 1 (`#542`) закладывает schema/backfill foundation и parity markers до любого domain, worker, transport или UI exposure; destructive cleanup и read-switch cleanup в этой wave запрещены.
- Wave 2 (`#543`) стартует только после подтверждённого foundation-evidence по `#542`; `control-plane` остаётся единственным owner graph truth, continuity gaps, workspace watermarks и launch preview semantics.
- Wave 3 (`#544`) запускается только после стабилизации `#543`; `worker` исполняет reconcile/warmup/backfill строго по persisted contracts `control-plane` и не выводит локально graph integrity, next-step policy или continuity semantics.
- Wave 4 (`#545`) открывает typed transport visibility только после readiness domain/reconcile foundations; `api-gateway` остаётся thin-edge adapter и не получает доменную логику или provider-side effects.
- Wave 5 (`#546`) запускается после `#545` и не дублирует policy logic во frontend; `web-console` работает только через typed projections и сохраняет graph-first workspace как primary UX.
- Wave 6 (`#547`) обязательна перед handover в `run:qa`: без observability/readiness evidence Sprint S16 не считается stage-ready даже при локально работающем graph flow.
- Sprint S9 dashboard-first baseline, voice/STT, dashboard orchestrator agent, отдельная `agent` node taxonomy, full-history/archive и richer provider enrichment остаются вне execution package Sprint S16.

## Quality gates (`run:plan`)

| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S16-D6-01` | Для всех execution streams созданы отдельные handover issues `#542..#547` | passed |
| `QG-S16-D6-02` | Sequencing-waves и rollout order зафиксированы в delivery-документации | passed |
| `QG-S16-D6-03` | `control-plane`/`worker`/`api-gateway`/`web-console` сохраняют Day4 ownership split без reopening graph truth boundaries | passed |
| `QG-S16-D6-04` | Day5 guardrails сохранены: `#480`, exact Wave 1 filters/nodes, secondary/dimmed only for integrity, read-only preview, no new deployable service | passed |
| `QG-S16-D6-05` | `api-gateway` и `web-console` остаются projection-driven surfaces без local policy inference | passed |
| `QG-S16-D6-06` | `#547` зафиксирован как обязательный readiness/observability gate перед `run:qa` | passed |
| `QG-S16-D6-07` | Traceability синхронизирована (`issue_map`, `delivery_plan`, sprint/epic docs, traceability history, indexes) | passed |
| `QG-S16-D6-08` | Scope этапа ограничен markdown-only изменениями и без auto-trigger labels на follow-up issues | passed |

## Definition of Ready / Done для перехода в `run:dev`

### Definition of Ready (`run:dev` launch)
- [x] Design package Day5 (`#519`) подтверждён как source of truth.
- [x] Execution backlog создан отдельными issues `#542..#547`.
- [x] Зафиксированы sequencing-waves и зависимости между schema/backfill foundation, graph truth, reconcile, transport, UI и readiness evidence.
- [x] Rollout order `migrations -> control-plane -> worker -> api-gateway -> web-console -> readiness gate` выражен явно.
- [x] Trigger-лейблы на implementation issues не выставлены автоматически.
- [x] Guardrails Sprint S16 и continuity rule `PR + linked follow-up issue` сохранены явно в handover backlog.

### Definition of Done (`run:plan` stage)
- [x] Выпущен plan-эпик Day6 с execution package, quality-gates и DoR/DoD.
- [x] Созданы handover issues `#542..#547` без trigger-лейблов.
- [x] Обновлены `delivery_plan`, sprint/epic каталоги, `issue_map`, history-пакет Sprint S16 и индексные README.
- [x] Зафиксированы blockers, risks и owner decisions для следующего stage.

## Self-check (common checklist)
- Проверен scope изменений: только markdown-документы (`*.md`), без кода, YAML/JSON, Dockerfile и shell-скриптов.
- Проверена консистентность stage-policy: `run:dev` остаётся единственным кодовым этапом, trigger-лейблы на follow-up issues не выставлялись.
- Проверена traceability-синхронизация для нового day6-эпика и issue refs `#542..#547`.
- Новые внешние зависимости не выбирались; dependency catalog не менялся.
- Секреты, токены и environment values в артефактах не публиковались.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | `BLK-S16-D6-01` | До старта `run:dev` нужен owner review/approval plan package по Issue `#537` | open |
| blocker | `BLK-S16-D6-02` | Trigger-лейблы `run:dev` на implementation issues `#542..#547` должен выставить Owner по wave-sequencing | open |
| risk | `RSK-S16-D6-01` | Если `#542` разрастётся beyond schema/backfill foundation, rollout order и parity evidence станут неоднозначными для всех следующих волн | monitoring |
| risk | `RSK-S16-D6-02` | Старт `#544` до стабилизации `#543` приведёт к drift между canonical graph truth и background reconcile execution | monitoring |
| risk | `RSK-S16-D6-03` | Если `#545` или `#546` начнут интерпретировать continuity semantics локально, thin-edge/projection boundary будет нарушен | monitoring |
| risk | `RSK-S16-D6-04` | Отставание `#547` по readiness evidence заблокирует `run:qa`, даже если core graph workspace станет локально доступен | monitoring |
| owner-decision | `OD-S16-D6-01` | Core rollout выполняется только по waves `#542 -> #543 -> #544 -> #545 -> #546 -> #547`; массовый параллельный старт запрещён | proposed |
| owner-decision | `OD-S16-D6-02` | `run:dev` triggers выставляются Owner по waves, без автоматического старта при создании follow-up issues | proposed |
| owner-decision | `OD-S16-D6-03` | Handover в `run:qa` допускается только после закрытия `#547` и подтверждённого readiness/observability evidence | proposed |
| owner-decision | `OD-S16-D6-04` | Sprint S16 не возвращается к Sprint S9 dashboard-first модели; voice/STT и richer provider enrichment остаются отдельным downstream контуром | proposed |

## Tooling validation
- Для неинтерактивного PR/issue flow синтаксис дополнительно сверен локально:
  - `gh issue create --help`
  - `gh pr create --help`
  - `gh pr edit --help`
- Через `gh issue create` оформлены handover issues `#542..#547`.
- Kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope.
- Новые внешние библиотеки на этапе Day6 не выбирались.

## Acceptance Criteria (Issue #537)
- [x] Подготовлен execution package минимум по waves `schema/backfill -> control-plane graph truth -> worker reconcile -> transport -> web-console -> readiness gate`.
- [x] Для каждой wave зафиксированы цель, scope, зависимости, expected artifacts, quality gates, DoR/DoD и owner-managed правило запуска следующей wave.
- [x] Day5 guardrails сохранены явно: `#480`, exact Wave 1 filters `open_only` и `assigned_to_me_or_unassigned`, active-state presets, secondary/dimmed only for integrity, nodes `discussion/work_item/run/pull_request`, existing command ledger, read-only launch preview и запрет на новый deployable сервис.
- [x] Обновлены delivery traceability-артефакты для перехода `run:plan -> run:dev`.
- [x] Подготовлены follow-up issues для `run:dev`: backlog `#542..#547` без trigger-лейблов.

## Handover в `run:dev`
- Следующий operational stage: `run:dev`.
- Implementation issues для запуска по waves: `#542..#547`.
- `#542` закладывает schema/backfill foundation, `#543` закрепляет `control-plane` graph truth и continuity projections, `#544` ограничен worker reconcile/freshness, `#545` публикует typed transport/preview surfaces, `#546` реализует graph-first workspace UX, `#547` остаётся обязательным readiness gate перед `run:qa`.
- Для каждого `run:dev` потока обязательны:
  - PR с проверками и evidence;
  - синхронное обновление traceability документов;
  - переход в `state:in-review` на PR и Issue после завершения итерации.
