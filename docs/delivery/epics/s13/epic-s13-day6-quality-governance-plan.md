---
doc_id: EPC-CK8S-S13-D6-QUALITY-GOVERNANCE
type: epic
title: "Epic S13 Day 6: Plan для Quality Governance System, execution waves и rollout gates (Issue #512)"
status: in-review
owner_role: EM
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [469, 471, 476, 484, 494, 512, 521, 522, 523, 524, 525]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-16-issue-512-plan-epic"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-16
---

# Epic S13 Day 6: Plan для Quality Governance System, execution waves и rollout gates (Issue #512)

## TL;DR
- Подготовлен execution package Sprint S13 для перехода в `run:dev` по `Quality Governance System`.
- Созданы отдельные handover issues `#521..#525` для foundation aggregate, worker feedback/backfill, staff transport + GitHub mirror, `web-console` visibility и финального readiness gate.
- Зафиксированы sequencing-waves, quality-gates, DoR/DoD и owner decisions для rollout order `migrations -> control-plane -> worker -> api-gateway -> web-console`.
- Сохранены design guardrails Sprint S13: hidden draft остаётся internal-only, `semantic wave map` остаётся первой publishable единицей, `high/critical` не допускают silent waivers, а `worker` не получает ownership canonical semantics.

## Контекст
- Stage continuity: `#469 -> #471 -> #476 -> #484 -> #494 -> #512`.
- Входной baseline: design package Day5
  - `docs/architecture/initiatives/s13_quality_governance_system/design_doc.md`
  - `docs/architecture/initiatives/s13_quality_governance_system/api_contract.md`
  - `docs/architecture/initiatives/s13_quality_governance_system/data_model.md`
  - `docs/architecture/initiatives/s13_quality_governance_system/migrations_policy.md`
- Scope текущего stage: только markdown-изменения и handover backlog.
- Правило continuity: trigger-лейблы `run:dev` на implementation issues ставит только Owner и только по wave-sequencing.

## Execution package (S13-E01..S13-E05)

| Stream | Implementation issue | Wave | Priority | Краткий scope |
|---|---:|---|---|---|
| `S13-E01` | #521 | Wave 1 | P0 | Additive schema, `control-plane` package aggregate, hidden draft / wave / evidence ingress и projection refresh под owner canonical semantics |
| `S13-E02` | #522 | Wave 2 | P0 | `worker` sweeps, governance-gap feedback, bounded historical backfill и late reclassification under `control-plane` policy |
| `S13-E03` | #523 | Wave 3 | P0 | Contract-first `api-gateway` staff/private transport surfaces и read-only GitHub status mirror |
| `S13-E04` | #524 | Wave 4 | P0 | `web-console` queue/detail/gap/release-readiness visibility на основе typed projections |
| `S13-E05` | #525 | Wave 5 | P0 | Observability, rollout/rollback discipline и readiness evidence gate перед `run:qa` |

## Sequencing constraints
- Wave 1 (`#521`) закладывает schema/domain foundation, typed hidden-draft / wave / evidence ingestion и projection versioning до любого `worker`, transport или UI exposure.
- Wave 2 (`#522`) стартует только после подтверждённого foundation-evidence по `#521`; `worker` исполняет sweeps/backfill только по persisted package/projection contracts и не изобретает собственную risk/evidence semantics.
- Wave 3 (`#523`) открывает transport visibility и GitHub mirror только после стабилизации package/projection model; `api-gateway` остаётся thin-edge, а comment mirror читает только persisted projections.
- Wave 4 (`#524`) запускается после `#523` и не дублирует policy logic во frontend; UI работает только через typed API contracts и отдельные projection states.
- Wave 5 (`#525`) обязательна перед handover в `run:qa`: без observability/readiness evidence capability не считается stage-ready даже при локально работающем core flow.
- `agent-runner` signal work остаётся bounded sub-scope Wave 1 и не меняет top-level rollout order `migrations -> control-plane -> worker -> api-gateway -> web-console`.
- Sprint S14 (`#470`) остаётся downstream runtime/UI stream: realtime/cockpit inventions не входят в execution package Sprint S13.

## Quality gates (`run:plan`)

| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S13-D6-01` | Для всех execution streams созданы отдельные handover issues `#521..#525` | passed |
| `QG-S13-D6-02` | Sequencing-waves и rollout order зафиксированы в delivery-документации | passed |
| `QG-S13-D6-03` | Hidden draft / semantic wave / waiver guardrails сохранены без reopening policy semantics | passed |
| `QG-S13-D6-04` | `worker` остаётся reconcile-only owner для sweeps/backfill и не получает canonical semantics package aggregate | passed |
| `QG-S13-D6-05` | `api-gateway` и `web-console` остаются projection-driven thin surfaces без local policy inference | passed |
| `QG-S13-D6-06` | `#525` зафиксирован как обязательный observability/readiness gate перед `run:qa` | passed |
| `QG-S13-D6-07` | Traceability синхронизирована (`issue_map`, `delivery_plan`, sprint/epic docs, traceability history) | passed |
| `QG-S13-D6-08` | Scope этапа ограничен markdown-only изменениями и без auto-trigger labels на follow-up issues | passed |

## Definition of Ready / Done для перехода в `run:dev`

### Definition of Ready (`run:dev` launch)
- [x] Design package Day5 (`#494`) подтверждён как source of truth.
- [x] Execution backlog создан отдельными issues `#521..#525`.
- [x] Зафиксированы sequencing-waves и зависимости между foundation, worker lifecycle, transport, UI и readiness evidence.
- [x] Rollout order `migrations -> control-plane -> worker -> api-gateway -> web-console` выражен явно.
- [x] Trigger-лейблы на implementation issues не выставлены автоматически.
- [x] Guardrails Sprint S13 сохранены явно в handover backlog.

### Definition of Done (`run:plan` stage)
- [x] Выпущен plan-эпик Day6 с execution package, quality-gates и DoR/DoD.
- [x] Созданы handover issues `#521..#525` без trigger-лейблов.
- [x] Обновлены `delivery_plan`, sprint/epic каталоги, `issue_map` и history-пакет Sprint S13.
- [x] Зафиксированы blockers, risks и owner decisions для следующего stage.

## Self-check (common checklist)
- Проверен scope изменений: только markdown-документы (`*.md`), без кода, YAML/JSON, Dockerfile и shell-скриптов.
- Проверена консистентность stage-policy: `run:dev` остаётся единственным кодовым этапом, trigger-лейблы на follow-up issues не выставлялись.
- Проверена traceability-синхронизация для нового day6-эпика и issue refs `#521..#525`.
- Новые внешние зависимости не выбирались; dependency catalog не менялся.
- Секреты, токены и environment values в артефактах не публиковались.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | `BLK-S13-D6-01` | До старта `run:dev` нужен owner review/approval plan package по Issue `#512` | open |
| blocker | `BLK-S13-D6-02` | Trigger-лейблы `run:dev` на implementation issues `#521..#525` должен выставить Owner по wave-sequencing | open |
| risk | `RSK-S13-D6-01` | Если `#521` разрастётся beyond foundation aggregate, rollout ambiguity и ownership drift заблокируют все последующие waves | monitoring |
| risk | `RSK-S13-D6-02` | Старт `#522` до стабилизации `#521` приведёт к drift между canonical semantics и background reconciliation | monitoring |
| risk | `RSK-S13-D6-03` | Если `#523` или `#524` начнут трактовать risk/evidence semantics локально, thin-edge/projection boundary будет нарушен | monitoring |
| risk | `RSK-S13-D6-04` | Отставание `#525` по readiness evidence заблокирует `run:qa`, даже если core functionality будет локально доступна | monitoring |
| owner-decision | `OD-S13-D6-01` | Core rollout выполняется только по waves `#521 -> #522 -> #523 -> #524 -> #525`; массовый параллельный старт запрещён | proposed |
| owner-decision | `OD-S13-D6-02` | `run:dev` triggers выставляются Owner по waves, без автоматического старта при создании follow-up issues | proposed |
| owner-decision | `OD-S13-D6-03` | Handover в `run:qa` допускается только после закрытия `#525` и подтверждённого observability/readiness evidence | proposed |
| owner-decision | `OD-S13-D6-04` | Sprint S14 (`#470`) остаётся downstream runtime/UI stream и не становится скрытым prerequisite core Sprint S13 rollout | proposed |

## Tooling validation
- Для неинтерактивного PR/issue flow синтаксис дополнительно сверен локально:
  - `gh issue create --help`
  - `gh pr create --help`
  - `gh pr edit --help`
- Через `gh issue create` оформлены handover issues `#521..#525`.
- Kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope.
- Новые внешние библиотеки на этапе Day6 не выбирались.

## Acceptance Criteria (Issue #512)
- [x] Подготовлен execution package по направлениям foundation aggregate, worker reconciliation/backfill, transport/mirror, UI visibility и observability/readiness.
- [x] Для каждого потока зафиксированы scope, зависимости, required checks, rollout order и expected success evidence.
- [x] Guardrails сохранены явно: hidden draft internal-only, `semantic wave map` mandatory, no silent waivers for `high/critical`, `worker` reconcile-only, thin-edge transport/UI boundaries.
- [x] Подготовлены follow-up issues для `run:dev`: backlog `#521..#525` без trigger-лейблов.

## Handover в `run:dev`
- Следующий operational stage: `run:dev`.
- Implementation issues для запуска по waves: `#521..#525`.
- `#521` закрепляет schema/domain foundation, `#522` ограничен worker feedback/backfill, `#523` публикует transport contracts и GitHub mirror, `#524` реализует UI visibility, `#525` остаётся обязательным evidence gate перед `run:qa`.
- Для каждого `run:dev` потока обязательны:
  - PR с проверками и evidence;
  - синхронное обновление traceability документов;
  - переход в `state:in-review` на PR и Issue после завершения итерации.
