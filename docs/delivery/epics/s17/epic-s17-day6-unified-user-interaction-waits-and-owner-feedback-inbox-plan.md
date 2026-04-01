---
doc_id: EPC-CK8S-S17-D6-OWNER-FEEDBACK-PLAN
type: epic
title: "Epic S17 Day 6: Plan для unified owner feedback loop, execution waves и handover в run:dev (Issue #575)"
status: in-review
owner_role: EM
created_at: 2026-04-01
updated_at: 2026-04-01
related_issues: [391, 392, 393, 394, 395, 458, 541, 554, 557, 559, 568, 575, 582]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-04-01-issue-575-plan-epic"
---

# Epic S17 Day 6: Plan для unified owner feedback loop, execution waves и handover в run:dev (Issue #575)

## TL;DR
- Подготовлен execution package Sprint S17 для перехода в `run:dev` по unified owner feedback loop и same-session continuation contract.
- Создана handover issue `#582` как единый execution anchor для `run:dev` с обязательным continuity-требованием сохранить цепочку `#541 -> #554 -> #557 -> #559 -> #568 -> #575 -> #582` без разрывов.
- Зафиксированы sequencing-waves, prerequisite gate на закрытых Sprint S10/S11 foundation issues `#391..#395` и `#458`, quality-gates, DoR/DoD и rollout order `migrations -> control-plane -> worker -> api-gateway -> telegram-interaction-adapter -> web-console -> observability/evidence gate`.
- Same-session happy-path, max timeout/TTL baseline, recovery-only snapshot-resume, dual-surface persisted truth и staff-console projection model сохранены как обязательные инварианты execution stage.

## Контекст
- Stage continuity: `#541 -> #554 -> #557 -> #559 -> #568 -> #575`.
- Входной baseline: design package Day5
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/design_doc.md`
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/api_contract.md`
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/data_model.md`
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/migrations_policy.md`
- Scope текущего stage: только markdown-изменения и handover backlog.
- Проверяемый foundation gate:
  - Sprint S10 implementation issues `#391`, `#392`, `#393`, `#394`, `#395` закрыты и сохраняют interaction foundation для built-in waits.
  - Sprint S11 execution anchor `#458` закрыта и сохраняет Telegram channel baseline для callback/delivery contour.
- Правило continuity: trigger-лейбл `run:dev` на issue `#582` ставит только Owner; plan package Issue `#575` фиксирует governance baseline для handover.

## Execution package (S17-E01..S17-E07)

| Stream | Execution anchor | Wave | Priority | Краткий scope | Expected success evidence |
|---|---:|---|---|---|---|
| `S17-E01` | #582 | Wave 1 | P0 | `control-plane` schema foundation: additive migrations, owner-feedback overlay поля и таблицы (`interaction_requests`, `owner_feedback_wait_links`, `owner_feedback_channel_projections`, `owner_feedback_response_bindings`, `interaction_response_records`, `agent_runs`) | Additive schema path, indexes и schema ownership подтверждены без drift относительно Sprint S10/S11 foundation |
| `S17-E02` | #582 | Wave 2 | P0 | `control-plane` domain/use-case path: persisted request truth, response binding registry, wait-state linkage, same-session vs recovery classification и typed resume payload | Один semantic winner на request, recovery остаётся explicit degraded path, а same-session continuation не деградирует в detached resume |
| `S17-E03` | #582 | Wave 3 | P0 | `worker` delivery/retry/reconcile/visibility path для `delivery_accepted`, `overdue`, `expired`, `manual_fallback` и `recovery_resume` | Visibility и background transitions исполняются только по persisted contracts `control-plane` без локальной worker-semantics |
| `S17-E04` | #582 | Wave 4 | P0 | Thin-edge `api-gateway` bridge: contract-first OpenAPI/codegen sync, typed DTO/casters, staff read/write endpoints и callback bridge | Transport surface публикует только typed contracts и не переносит lifecycle semantics в edge layer |
| `S17-E05` | #582 | Wave 5 | P0 | `telegram-interaction-adapter` delivery/callback/voice normalization path, opaque handle verification и provider evidence bridge | Adapter contour остаётся только Bot API/normalization/auth bridge и не выбирает final response winner |
| `S17-E06` | #582 | Wave 6 | P0 | `staff web-console` inbox/fallback UX, projection-driven pending list/detail, typed response submission и visibility rendering | UI остаётся projection-driven surface с backend-derived `allowed_actions` и не становится вторым source of truth |
| `S17-E07` | #582 | Wave 7 | P0 | Observability, acceptance evidence, rollout/rollback gates и readiness handover перед `run:qa` | Есть candidate evidence по same-session, recovery-only, manual fallback, rollout discipline и dual-surface visibility |

## Sequencing constraints
- Wave 1 (`S17-E01`) стартует только при сохранении prerequisite Sprint S10/S11 foundation: issues `#391..#395` и `#458` остаются закрытыми и не теряют значение source of truth для current interaction/Telegram baseline.
- Wave 2 (`S17-E02`) стартует только после подтверждённого foundation-evidence по Wave 1; `control-plane` остаётся единственным owner request truth, winner selection, continuation classification и typed resume payload.
- Wave 3 (`S17-E03`) запускается только после завершения Wave 2; `worker` исполняет delivery/retry/reconcile path строго по persisted domain contracts и не выводит локально canonical status.
- Wave 4 (`S17-E04`) открывает transport visibility только после стабилизации Wave 3; `api-gateway` остаётся thin-edge bridge и не получает ad hoc бизнес-логики owner feedback lifecycle.
- Wave 5 (`S17-E05`) стартует только после готовности Wave 4; Telegram adapter contour не может обойти platform-owned contracts прямым webhook-to-domain shortcut.
- Wave 6 (`S17-E06`) запускается только после стабилизации Wave 5 и использует тот же response binding registry/allowed-actions matrix; staff-console не может получить shortcut write path вне typed backend contract.
- Wave 7 (`S17-E07`) обязательна перед handover в `run:qa` и фиксирует observability, fallback/manual-action readiness и acceptance evidence для всего owner feedback loop.
- Дополнительные каналы, reminders/escalations, attachments, multi-party routing, generalized conversation UX и detached resume-run как равноправный happy-path не входят в execution package Sprint S17.

## Quality gates (`run:plan`)

| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S17-D6-01` | Создана handover issue `#582` для `run:dev` с continuity-требованием `#541 -> #554 -> #557 -> #559 -> #568 -> #575 -> #582` | passed |
| `QG-S17-D6-02` | Prerequisite foundation подтверждён: issues `#391..#395` и `#458` закрыты и не конфликтуют с S17 handover | passed |
| `QG-S17-D6-03` | Sequencing-waves и rollout order выражены явно в delivery-документации | passed |
| `QG-S17-D6-04` | Ownership split сохранён: `control-plane` / `worker` / `api-gateway` / `telegram-interaction-adapter` / `web-console` не смешивают доменные обязанности | passed |
| `QG-S17-D6-05` | Same-session primary path, max timeout/TTL baseline и recovery-only snapshot-resume сохранены явно | passed |
| `QG-S17-D6-06` | One persisted truth и response binding registry сохранены: staff-console остаётся projection, а Telegram не становится owner semantics | passed |
| `QG-S17-D6-07` | Issue `#582` зафиксирована как единственный execution anchor без автоматической постановки trigger-лейбла | passed |
| `QG-S17-D6-08` | Observability, acceptance evidence и rollout/rollback gate зафиксированы как обязательные перед `run:qa` | passed |
| `QG-S17-D6-09` | Traceability синхронизирована (`delivery_plan`, sprint/epic docs, indexes, `issue_map`, history bundle, `requirements_traceability`) и scope этапа остался markdown-only | passed |

## Definition of Ready / Done для перехода в `run:dev`

### Definition of Ready (`run:dev` launch)
- [x] Design package Day5 (`#568`) подтверждён как source of truth.
- [x] Execution anchor `#582` создан без trigger-лейбла.
- [x] Зафиксированы sequencing-waves между schema foundation, domain semantics, worker visibility, transport bridge, Telegram contour, staff-console UX и observability evidence.
- [x] Rollout order `migrations -> control-plane -> worker -> api-gateway -> telegram-interaction-adapter -> web-console -> observability/evidence gate` выражен явно.
- [x] Continuity-цепочка `#541 -> #554 -> #557 -> #559 -> #568 -> #575 -> #582` повторно зафиксирована как обязательная.
- [x] Foundation prerequisites `#391..#395` и `#458` подтверждены как закрытые.

### Definition of Done (`run:plan` stage)
- [x] Выпущен plan-эпик Day6 с execution package, quality-gates и DoR/DoD.
- [x] Создана handover issue `#582` без trigger-лейбла.
- [x] Обновлены `delivery_plan`, sprint/epic каталоги, индексы, `issue_map`, `requirements_traceability` и history-пакет Sprint S17.
- [x] Зафиксированы blockers, risks и owner decisions для следующего stage.

## Self-check (common checklist)
- Проверен scope изменений: только markdown-документы (`*.md`), без кода, YAML/JSON, Dockerfile и shell-скриптов.
- Проверена консистентность stage-policy: `run:dev` остаётся единственным кодовым этапом, trigger-лейбл на issue `#582` автоматически не выставлялся.
- Проверена traceability-синхронизация для нового day6-эпика и issue refs `#575` / `#582`.
- Новые внешние зависимости и версии на этапе Day6 не выбирались; dependency catalog не менялся.
- Секреты, токены и environment values в артефактах не публиковались.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | `BLK-S17-D6-01` | Owner review/approval plan package по Issue `#575` требуется до старта `run:dev` | open |
| blocker | `BLK-S17-D6-02` | Trigger-лейбл `run:dev` на issue `#582` должен поставить Owner после review sequencing-waves и quality-gates | open |
| risk | `RSK-S17-D6-01` | Если Wave 1 разрастётся beyond additive owner-feedback overlay, ownership drift по `control-plane` заблокирует все последующие waves | monitoring |
| risk | `RSK-S17-D6-02` | Старт Worker/edge/UI waves до стабилизации Wave 2 приведёт к drift между canonical request truth и surface-level visibility | monitoring |
| risk | `RSK-S17-D6-03` | Если `api-gateway`, Telegram adapter или `web-console` начнут интерпретировать lifecycle semantics локально, dual-surface truth будет размыта | monitoring |
| risk | `RSK-S17-D6-04` | Любой rollback effective wait timeout/TTL ниже owner wait window сломает primary same-session model и нормализует degraded continuation | monitoring |
| risk | `RSK-S17-D6-05` | Отставание Wave 7 по observability/evidence заблокирует `run:qa`, даже если локально wait/respond path уже работает | monitoring |
| owner-decision | `OD-S17-D6-01` | Реализация идёт только по issue `#582` и по последовательным waves `S17-E01 -> S17-E02 -> S17-E03 -> S17-E04 -> S17-E05 -> S17-E06 -> S17-E07`; параллельный execution anchor не создаётся | accepted |
| owner-decision | `OD-S17-D6-02` | `run:dev` trigger на issue `#582` выставляется только Owner и только после review/approval plan package Issue `#575` | accepted |
| owner-decision | `OD-S17-D6-03` | Handover в `run:qa` допускается только после acceptance evidence Wave 7 и подтверждённого rollback/manual-fallback readiness | accepted |
| owner-decision | `OD-S17-D6-04` | Дополнительные каналы, reminders/escalations, attachments, multi-party routing и generalized conversation UX остаются отдельным follow-up контуром вне core Sprint S17 | accepted |

## Tooling validation
- Для неинтерактивного PR/issue flow синтаксис дополнительно сверен локально:
  - `gh issue create --help`
  - `gh issue edit --help`
  - `gh issue view --help`
  - `gh pr create --help`
  - `gh pr edit --help`
  - `gh pr view --help`
- Через `gh issue view` дополнительно подтверждён prerequisite gate:
  - `gh issue view 391 --json number,title,state,url`
  - `gh issue view 392 --json number,title,state,url`
  - `gh issue view 393 --json number,title,state,url`
  - `gh issue view 394 --json number,title,state,url`
  - `gh issue view 395 --json number,title,state,url`
  - `gh issue view 458 --json number,title,state,url`
- Через `gh issue create` оформлена handover issue `#582`.
- Kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope.
- Новые внешние библиотеки на этапе Day6 не выбирались.

## Acceptance Criteria (Issue #575)
- [x] Подготовлен execution package по потокам schema ownership, domain/use-case, worker visibility, thin-edge transport, Telegram adapter contour, staff web-console fallback и observability/evidence gate.
- [x] Для каждого потока зафиксированы scope, зависимости, required checks, rollout order и expected success evidence.
- [x] Guardrails сохранены явно: same-session happy-path, max timeout/TTL baseline, recovery-only snapshot-resume, dual-surface persisted truth, typed transport boundary и owner-managed review gate.
- [x] Подготовлена follow-up issue `#582` для `run:dev`, где явно повторено continuity-требование продолжить цепочку без разрывов.

## Handover в `run:dev`
- Следующий operational stage: `run:dev`.
- Execution anchor для запуска: issue `#582`.
- Wave order для `#582`: `S17-E01 -> S17-E02 -> S17-E03 -> S17-E04 -> S17-E05 -> S17-E06 -> S17-E07`.
- Для `run:dev` обязательны:
  - PR с проверками и evidence;
  - синхронное обновление traceability документов;
  - переход в `state:in-review` на PR и Issue после завершения итерации;
  - сохранение issue-цепочки `#541 -> #554 -> #557 -> #559 -> #568 -> #575 -> #582` без разрывов.
