---
doc_id: EPC-CK8S-S10-D6-MCP-INTERACTIONS
type: epic
title: "Epic S10 Day 6: Plan для built-in MCP user interactions (Issue #389)"
status: in-review
owner_role: EM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387, 389, 391, 392, 393, 394, 395]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-389-plan"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic S10 Day 6: Plan для built-in MCP user interactions (Issue #389)

## TL;DR
- Подготовлен execution package Sprint S10 для перехода в `run:dev` по built-in MCP user interactions.
- Созданы отдельные handover issues `#391..#395` для `control-plane` foundation, worker dispatch/retry/expiry, contract-first callback ingress, deterministic resume path в `agent-runner` и observability/readiness.
- Зафиксированы sequencing-waves, quality-gates, DoR/DoD и owner decisions для core rollout `migrations -> control-plane -> worker -> api-gateway`, а `agent-runner` resume path вынесен в отдельный bounded stream.
- Channel-specific adapters и Telegram остаются за пределами core Sprint S10 execution package и не блокируют handover в `run:dev`.

## Контекст
- Stage continuity: `#360 -> #378 -> #383 -> #385 -> #387 -> #389`.
- Входной baseline: design package Day5
  - `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`
  - `docs/architecture/initiatives/s10_mcp_user_interactions/data_model.md`
  - `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`
- Scope текущего stage: только markdown-изменения и handover backlog.
- Правило continuity: trigger-лейблы `run:dev` на implementation issues ставит только Owner и только по wave-sequencing.

## Execution package (S10-E01..S10-E05)

| Stream | Implementation issue | Wave | Priority | Краткий scope |
|---|---:|---|---|---|
| `S10-E01` | #391 | Wave 1 | P0 | `control-plane` foundation: additive schema, built-in tool orchestration, interaction aggregate, callback classification и typed wait linkage |
| `S10-E02` | #392 | Wave 2 | P0 | `worker` dispatch, retries, expiry scans, delivery-attempt ledger и resume scheduling |
| `S10-E03` | #393 | Wave 3 | P0 | Contract-first `api-gateway` callback ingress, OpenAPI/codegen sync, typed DTO/casters и thin-edge auth/error mapping |
| `S10-E04` | #394 | Wave 3 | P0 | `agent-runner` deterministic resume path, typed payload handoff и resume prompt block |
| `S10-E05` | #395 | Wave 4 | P0 | Observability, replay/idempotency evidence, rollout/rollback gates и acceptance-readiness перед `run:qa` |

## Sequencing constraints
- Wave 1 (`#391`) закладывает schema/domain foundation, interaction aggregate и typed wait linkage до любого dispatch/callback exposure.
- Wave 2 (`#392`) стартует только после подтверждённого foundation-evidence по `#391` и не дублирует callback classification/business semantics.
- Wave 3 (`#393`, `#394`) допускает ограниченный параллелизм только после завершения `#392`:
  - `#393` реализует только thin-edge callback ingress и gRPC bridge на уже зафиксированном domain contract;
  - `#394` интегрирует deterministic resume path только против persisted `interaction_resume_payload` и existing `agent_sessions` snapshot path.
- Открытие callback path и end-to-end resume flow разрешается только после готовности `#391` + `#392`; `#393` и `#394` не могут обойти эти gates локальными допущениями.
- Wave 4 (`#395`) обязательна перед handover в `run:qa` и фиксирует evidence по replay/idempotency, retry backlog, expiry/resume correctness и rollout discipline.
- Telegram/adapters, richer conversations, reminders и voice/STT не входят в этот execution package и не становятся скрытым prerequisite core rollout.

## Quality gates (`run:plan`)

| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S10-D6-01` | Для всех execution streams созданы отдельные handover issues `#391..#395` | passed |
| `QG-S10-D6-02` | Sequencing-waves и зависимости зафиксированы в delivery-документации | passed |
| `QG-S10-D6-03` | Ownership split сохранён: `control-plane`/`worker`/`api-gateway`/`agent-runner` не смешивают доменные обязанности | passed |
| `QG-S10-D6-04` | Rollout order `migrations -> control-plane -> worker -> api-gateway` и отдельный resume gate для `agent-runner` выражены явно | passed |
| `QG-S10-D6-05` | Replay/idempotency/expiry/resume evidence зафиксированы как обязательный gate перед `run:qa` | passed |
| `QG-S10-D6-06` | Traceability синхронизирована (`issue_map`, delivery plan, sprint/epic docs, traceability history, indexes) | passed |
| `QG-S10-D6-07` | Scope этапа ограничен markdown-only изменениями и без auto-trigger labels на follow-up issues | passed |

## Definition of Ready / Done для перехода в `run:dev`

### Definition of Ready (`run:dev` launch)
- [x] Design package Day5 (`#387`) подтверждён как source of truth.
- [x] Execution backlog создан отдельными issues `#391..#395`.
- [x] Зафиксированы sequencing-waves и зависимости между foundation, worker lifecycle, callback transport, resume path и observability.
- [x] Replay/idempotency/resume gates выражены явно и не размыты между сервисными границами.
- [x] Trigger-лейблы на implementation issues не выставлены автоматически.

### Definition of Done (`run:plan` stage)
- [x] Выпущен plan-эпик Day6 с execution package, QG и DoR/DoD.
- [x] Созданы handover issues `#391..#395` без trigger-лейблов.
- [x] Обновлены `delivery_plan`, sprint/epic каталоги, `issue_map` и history-пакет Sprint S10.
- [x] Зафиксированы blockers, risks и owner decisions для следующего stage.

## Self-check (common checklist)
- Проверен scope изменений: только markdown-документы (`*.md`), без кода, YAML/JSON, Dockerfile и shell-скриптов.
- Проверена консистентность stage-policy: `run:dev` остаётся единственным кодовым этапом, trigger-лейблы на follow-up issues не выставлялись.
- Проверена traceability-синхронизация для нового day6-эпика и issue refs `#391..#395`.
- Новые внешние зависимости не выбирались; dependency catalog не менялся.
- Секреты, токены и environment values в артефактах не публиковались.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | `BLK-S10-D6-01` | До старта `run:dev` нужен owner review/approval plan package по Issue `#389` | open |
| blocker | `BLK-S10-D6-02` | Trigger-лейблы `run:dev` на implementation issues `#391..#395` должен выставить Owner по wave-sequencing | open |
| risk | `RSK-S10-D6-01` | Если `#391` разрастётся beyond `control-plane` foundation, ownership drift снова смешает interaction и approval semantics | monitoring |
| risk | `RSK-S10-D6-02` | Параллельный старт `#393` или `#394` до стабилизации `#392` может дать callback/resume drift и неконсистентный rollout | monitoring |
| risk | `RSK-S10-D6-03` | Если `#394` выйдет за рамки typed interaction handoff, можно случайно переоткрыть общий pause/resume engine без отдельного design решения | monitoring |
| risk | `RSK-S10-D6-04` | Отставание `#395` по evidence заблокирует `run:qa`, даже если core functionality будет локально работать | monitoring |
| owner-decision | `OD-S10-D6-01` | Core rollout выполняется только по issues `#391 -> #392 -> #393/#394 -> #395`; массовый параллельный старт всех streams запрещён | proposed |
| owner-decision | `OD-S10-D6-02` | `run:dev` triggers выставляются Owner по waves, без автоматического старта при создании follow-up issues | proposed |
| owner-decision | `OD-S10-D6-03` | Handover в `run:qa` допускается только после закрытия `#395` и подтверждённого replay/idempotency/resume evidence | proposed |
| owner-decision | `OD-S10-D6-04` | Channel-specific adapters, Telegram, reminders и voice/STT остаются отдельным follow-up контуром после core Sprint S10 | proposed |

## Tooling validation
- Попытка использовать Context7 для GitHub CLI manual завершилась ошибкой `Monthly quota exceeded`.
- Для неинтерактивного PR/issue flow синтаксис дополнительно сверен локально:
  - `gh issue create --help`
  - `gh pr create --help`
  - `gh pr edit --help`
- Новые внешние библиотеки на этапе Day6 не выбирались.

## Acceptance Criteria (Issue #389)
- [x] Подготовлен execution package по потокам `control-plane`, `worker`, `api-gateway`, `agent-runner` и observability/quality readiness.
- [x] Для каждого потока зафиксированы scope, зависимости, required checks, rollout order и expected success evidence.
- [x] Guardrails сохранены явно: separation from approval flow, thin-edge boundary, deterministic resume payload, replay/idempotency safety и owner-managed review gate.
- [x] Подготовлены follow-up issues для `run:dev`: core backlog `#391..#395` без trigger-лейблов.

## Handover в `run:dev`
- Следующий operational stage: `run:dev`.
- Implementation issues для запуска по waves: `#391..#395`.
- `#391` закрепляет schema/domain foundation, `#392` владеет dispatch/retry/expiry, `#393` ограничен callback transport, `#394` закрывает deterministic resume path, `#395` остаётся обязательным evidence gate перед `run:qa`.
- Для каждого `run:dev` потока обязательны:
  - PR с проверками и evidence;
  - синхронное обновление traceability документов;
  - переход в `state:in-review` на PR и Issue после завершения итерации.
