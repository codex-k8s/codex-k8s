---
doc_id: EPC-CK8S-S12-D6-GITHUB-RATE-LIMIT
type: epic
title: "Epic S12 Day 6: Plan для GitHub API rate-limit resilience, execution waves и rollout gates (Issue #423)"
status: completed
owner_role: EM
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-issue-423-plan-epic"
---

# Epic S12 Day 6: Plan для GitHub API rate-limit resilience, execution waves и rollout gates (Issue #423)

## TL;DR
- Подготовлен execution package Sprint S12 для перехода в `run:dev` по GitHub API rate-limit resilience.
- Созданы отдельные handover issues `#425..#431` для schema foundation, `control-plane`, `worker`, `agent-runner`, `api-gateway`, `web-console` и observability/readiness gate.
- Зафиксированы sequencing-waves, quality-gates, DoR/DoD и owner decisions для rollout order `migrations -> control-plane -> worker -> agent-runner -> api-gateway -> web-console -> evidence gate`.
- GitHub-first baseline, split `platform PAT` vs `agent bot-token`, hard-failure separation и запрет infinite local retries сохранены как обязательные инварианты execution stage.

## Контекст
- Stage continuity: `#366 -> #413 -> #416 -> #418 -> #420 -> #423`.
- Входной baseline: design package Day5
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/design_doc.md`
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/api_contract.md`
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/data_model.md`
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/migrations_policy.md`
- Scope текущего stage: только markdown-изменения и handover backlog.
- Правило continuity: trigger-лейблы `run:dev` на implementation issues ставит только Owner и только по wave-sequencing.

## Execution package (S12-E01..S12-E07)

| Stream | Implementation issue | Wave | Priority | Краткий scope |
|---|---:|---|---|---|
| `S12-E01` | #425 | Wave 1 | P0 | Wait-state persistence, additive schema, evidence ledger и dominant wait linkage under `control-plane` ownership |
| `S12-E02` | #426 | Wave 2 | P0 | `control-plane` classification, contour attribution, visibility projection, manual guidance и resume payload semantics |
| `S12-E03` | #427 | Wave 3 | P0 | `worker` auto-resume sweeps, bounded retries, replay scheduling и manual-action escalation |
| `S12-E04` | #428 | Wave 4 | P0 | `agent-runner` rate-limit handoff, persisted session snapshot и deterministic resume payload consumption |
| `S12-E05` | #429 | Wave 5 | P0 | Contract-first `api-gateway` visibility contracts, DTO/casters и realtime transport exposure |
| `S12-E06` | #430 | Wave 6 | P0 | `web-console` wait queue, run visibility, contour attribution и typed manual-action UX |
| `S12-E07` | #431 | Wave 7 | P0 | Observability, rollout/rollback discipline и readiness evidence gate before `run:qa` |

## Sequencing constraints
- Wave 1 (`#425`) закладывает schema/index/repository foundation и dominant wait linkage до любого resume/visibility exposure.
- Wave 2 (`#426`) стартует только после подтверждённого foundation-evidence по `#425` и остаётся единственным owner для classification, recovery hints и canonical wait projection.
- Wave 3 (`#427`) запускается только после завершения `#426`; `worker` исполняет time-based orchestration строго по persisted contracts и не изобретает свою классификацию rate-limit сигналов.
- Wave 4 (`#428`) запускается только после стабилизации `#427`, чтобы `agent-runner` интегрировался с уже зафиксированными wait/resume semantics и не кодировал stale retry behavior локально.
- Wave 5 (`#429`) открывает transport visibility только после готовности `#428`; `api-gateway` публикует typed projection и realtime events без доменных решений inside handlers.
- Wave 6 (`#430`) запускается после `#429` и не дублирует classification/recovery policy во frontend; UI работает только через typed API contracts.
- Wave 7 (`#431`) обязательна перед handover в `run:qa` и фиксирует observability/readiness evidence, rollout order и rollback constraints для всей capability.
- Predictive budgeting, multi-provider governance и adapter-specific notifications не входят в execution package Sprint S12 и не становятся скрытым prerequisite core rollout.

## Quality gates (`run:plan`)

| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S12-D6-01` | Для всех execution streams созданы отдельные handover issues `#425..#431` | passed |
| `QG-S12-D6-02` | Sequencing-waves и rollout order выражены явно в delivery-документации | passed |
| `QG-S12-D6-03` | Ownership split сохранён: `control-plane`/`worker`/`agent-runner`/`api-gateway`/`web-console` не смешивают доменные обязанности | passed |
| `QG-S12-D6-04` | GitHub-first guardrails сохранены: split contours, hard-failure separation и no infinite local retries | passed |
| `QG-S12-D6-05` | Thin-edge visibility path зафиксирован как typed projection без UI/transport inference from raw logs or headers | passed |
| `QG-S12-D6-06` | `#431` зафиксирован как обязательный observability/readiness gate перед `run:qa` | passed |
| `QG-S12-D6-07` | Traceability синхронизирована (`issue_map`, delivery plan, sprint/epic docs, traceability history) | passed |
| `QG-S12-D6-08` | Scope этапа ограничен markdown-only изменениями и без auto-trigger labels на follow-up issues | passed |

## Definition of Ready / Done для перехода в `run:dev`

### Definition of Ready (`run:dev` launch)
- [x] Design package Day5 (`#420`) подтверждён как source of truth.
- [x] Execution backlog создан отдельными issues `#425..#431`.
- [x] Зафиксированы sequencing-waves и зависимости между persistence, domain semantics, worker orchestration, runner handoff, transport, UI и observability.
- [x] Rollout order `migrations -> control-plane -> worker -> agent-runner -> api-gateway -> web-console -> evidence gate` выражен явно.
- [x] Trigger-лейблы на implementation issues не выставлены автоматически.

### Definition of Done (`run:plan` stage)
- [x] Выпущен plan-эпик Day6 с execution package, quality-gates и DoR/DoD.
- [x] Созданы handover issues `#425..#431` без trigger-лейблов.
- [x] Обновлены `delivery_plan`, sprint/epic каталоги, `issue_map` и history-пакет Sprint S12.
- [x] Зафиксированы blockers, risks и owner decisions для следующего stage.

## Self-check (common checklist)
- Проверен scope изменений: только markdown-документы (`*.md`), без кода, YAML/JSON, Dockerfile и shell-скриптов.
- Проверена консистентность stage-policy: `run:dev` остаётся единственным кодовым этапом, trigger-лейблы на follow-up issues не выставлялись.
- Проверена traceability-синхронизация для нового day6-эпика и issue refs `#425..#431`.
- Новые внешние зависимости не выбирались; dependency catalog не менялся.
- Секреты, токены и environment values в артефактах не публиковались.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | `BLK-S12-D6-01` | До старта `run:dev` нужен owner review/approval plan package по Issue `#423` | open |
| blocker | `BLK-S12-D6-02` | Trigger-лейблы `run:dev` на implementation issues `#425..#431` должен выставить Owner по wave-sequencing | open |
| risk | `RSK-S12-D6-01` | Если `#425` разрастётся beyond schema foundation, shared-schema drift и rollout ambiguity заблокируют все последующие waves | monitoring |
| risk | `RSK-S12-D6-02` | Старт `#427` до стабилизации `#426` приведёт к retry/resume drift между domain semantics и worker orchestration | monitoring |
| risk | `RSK-S12-D6-03` | Если `#428` начнёт кодировать retry semantics локально, agent path нарушит guardrail no infinite local retries | monitoring |
| risk | `RSK-S12-D6-04` | Если `#429` или `#430` начнут выводить смысл из raw logs/service-comment, visibility drift разрушит controlled wait UX | monitoring |
| risk | `RSK-S12-D6-05` | Отставание `#431` по observability/readiness evidence заблокирует `run:qa` даже при локально рабочей функциональности | monitoring |
| owner-decision | `OD-S12-D6-01` | Core rollout выполняется только по waves `#425 -> #426 -> #427 -> #428 -> #429 -> #430 -> #431`; массовый параллельный старт запрещён | proposed |
| owner-decision | `OD-S12-D6-02` | `run:dev` triggers выставляются Owner по waves, без автоматического старта при создании follow-up issues | proposed |
| owner-decision | `OD-S12-D6-03` | Handover в `run:qa` допускается только после закрытия `#431` и подтверждённого observability/readiness evidence | proposed |
| owner-decision | `OD-S12-D6-04` | GitHub-first baseline сохраняется; predictive budgeting, multi-provider governance и adapter-specific notifications остаются отдельным follow-up контуром | proposed |

## Tooling validation
- Context7 `/github/docs` использован для актуальной верификации guidance GitHub REST API:
  - primary vs secondary rate-limit semantics;
  - приоритет `Retry-After`, `x-ratelimit-remaining` и `x-ratelimit-reset`;
  - рекомендация ждать минимум минуту и применять exponential backoff при отсутствии `Retry-After`;
  - avoidance of concurrency bursts.
- Для неинтерактивного PR/issue flow синтаксис дополнительно сверен локально:
  - `gh issue create --help`
  - `gh pr create --help`
  - `gh pr edit --help`
- Новые внешние библиотеки на этапе Day6 не выбирались.

## Acceptance Criteria (Issue #423)
- [x] Подготовлен execution package минимум по направлениям persistence, `control-plane`, `worker`, `agent-runner`, visibility surfaces и observability/readiness.
- [x] Для каждого потока зафиксированы scope, зависимости, required checks, rollout order и expected success evidence.
- [x] Guardrails сохранены явно: GitHub-first baseline, split contours, hard-failure separation, thin-edge boundaries и no infinite local retries.
- [x] Подготовлены follow-up issues для `run:dev`: backlog `#425..#431` без trigger-лейблов.

## Handover в `run:dev`
- Следующий operational stage: `run:dev`.
- Implementation issues для запуска по waves: `#425..#431`.
- `#425` закладывает persistence foundation, `#426` владеет classification/projection, `#427` закрывает worker auto-resume, `#428` ограничен runner handoff/resume path, `#429` публикует transport contracts, `#430` реализует UI transparency, `#431` остаётся обязательным evidence gate перед `run:qa`.
- Для каждого `run:dev` потока обязательны:
  - PR с проверками и evidence;
  - синхронное обновление traceability документов;
  - переход в `state:in-review` на PR и Issue после завершения итерации.
