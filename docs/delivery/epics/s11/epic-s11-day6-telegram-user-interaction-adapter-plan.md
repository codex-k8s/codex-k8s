---
doc_id: EPC-CK8S-S11-D6-TELEGRAM-ADAPTER
type: epic
title: "Epic S11 Day 6: Plan для Telegram-адаптера взаимодействия с пользователем, sequencing-waves и handover в run:dev (Issue #456)"
status: completed
owner_role: EM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456, 458]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-456-plan-epic"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# Epic S11 Day 6: Plan для Telegram-адаптера взаимодействия с пользователем, sequencing-waves и handover в run:dev (Issue #456)

## TL;DR
- Подготовлен execution package Sprint S11 для перехода в `run:dev` по Telegram-адаптеру взаимодействия с пользователем.
- Создана handover issue `#458` как единый execution anchor для `run:dev` с обязательным требованием продолжить issue-цепочку `#361 -> #447 -> #448 -> #452 -> #454 -> #456 -> #458` без разрывов.
- Зафиксированы sequencing-waves, quality-gates, DoR/DoD, blockers и owner decisions для rollout order `migrations -> control-plane -> worker -> api-gateway -> Telegram adapter contour -> observability/evidence gate`.
- Platform-owned semantics, separation from approval flow, typed transport boundary и dependency gate на Sprint S10 interaction foundation сохранены как обязательные инварианты execution stage.

## Контекст
- Stage continuity: `#361 -> #447 -> #448 -> #452 -> #454 -> #456`.
- Входной baseline: design package Day5
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/design_doc.md`
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/api_contract.md`
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/data_model.md`
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/migrations_policy.md`
- Scope текущего stage: только markdown-изменения и handover backlog.
- Правило continuity: trigger-лейбл `run:dev` на issue `#458` ставит только Owner; plan package Issue `#456` зафиксирован как завершённый governance baseline для этого handover.

## Execution package (S11-E01..S11-E06)

| Stream | Execution anchor | Wave | Priority | Краткий scope | Expected success evidence |
|---|---:|---|---|---|---|
| `S11-E01` | #458 | Wave 1 | P0 | `control-plane` schema foundation: additive migrations, `interaction_channel_bindings`, `interaction_callback_handles`, callback token lifecycle и repository surface под Telegram bindings | Миграции, repository contracts и additive schema path подтверждены без дрейфа ownership |
| `S11-E02` | #458 | Wave 2 | P0 | `control-plane` domain/use-case path: callback classification, response intake, operator visibility, canonical continuation semantics и typed outbound/inbound orchestration | Domain semantics остаются platform-owned и детерминированно проецируют duplicate/stale/expired outcomes |
| `S11-E03` | #458 | Wave 3 | P0 | `worker` delivery/retry/expiry/edit-follow-up path для notify/decision/free-text continuation и manual fallback evidence | Retry/expiry/edit-follow-up flow использует только persisted contracts и не вводит локальную семантику callback |
| `S11-E04` | #458 | Wave 4 | P0 | Thin-edge `api-gateway` bridge: contract-first OpenAPI/gRPC changes, typed DTO/casters и error mapping только на transport boundary | Transport surface публикует только typed contracts и не переносит доменную логику в edge layer |
| `S11-E05` | #458 | Wave 5 | P0 | Telegram adapter contour integration: raw webhook/auth, callback acknowledgement, provider refs и rollout coordination без Telegram-first semantics | Adapter contour замыкает только Bot API coupling, secret-token verify и callback UX acknowledgement |
| `S11-E06` | #458 | Wave 6 | P0 | Observability, fallback и acceptance evidence gate перед `run:qa` | Есть evidence по rollout/fallback/manual action readiness и наблюдаемости candidate path |

## Sequencing constraints
- Wave 1 (`S11-E01`) стартует только при сохранении prerequisite Sprint S10: Issue `#389` остаётся closed и design package Issue `#387` продолжает быть source of truth для interaction foundation.
- Wave 2 (`S11-E02`) стартует только после подтверждённого foundation-evidence по Wave 1; `control-plane` остаётся единственным owner callback classification, response intake и canonical continuation semantics.
- Wave 3 (`S11-E03`) запускается только после завершения Wave 2; `worker` исполняет delivery/retry/expiry/edit-follow-up path строго по persisted domain contracts и не изобретает свою классификацию callback outcomes.
- Wave 4 (`S11-E04`) открывает transport visibility только после стабилизации Wave 3; `api-gateway` остаётся thin-edge bridge и публикует только typed projection без локальных решений по Telegram semantics.
- Wave 5 (`S11-E05`) стартует только после готовности Wave 4; Telegram adapter contour не может обойти platform-owned contracts прямым webhook-to-domain shortcut.
- Wave 6 (`S11-E06`) обязательна перед handover в `run:qa` и фиксирует observability, fallback и acceptance evidence для всего Telegram path.
- Voice/STT, reminders, richer conversation threads, multi-chat routing policy и дополнительные каналы не входят в execution package Sprint S11 и не становятся скрытым prerequisite core rollout.

## Quality gates (`run:plan`)

| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S11-D6-01` | Создана handover issue `#458` для `run:dev` с явным continuity-требованием `#361 -> #447 -> #448 -> #452 -> #454 -> #456 -> #458` | passed |
| `QG-S11-D6-02` | Sequencing-waves и rollout order выражены явно в delivery-документации | passed |
| `QG-S11-D6-03` | Ownership split сохранён: `control-plane` / `worker` / `api-gateway` / Telegram adapter contour не смешивают доменные обязанности | passed |
| `QG-S11-D6-04` | Platform-owned semantics, separation from approval flow и typed transport boundary сохранены явно | passed |
| `QG-S11-D6-05` | `#458` зафиксирован как единый execution anchor без автоматической постановки trigger-лейбла | passed |
| `QG-S11-D6-06` | Observability/fallback evidence gate зафиксирован как обязательный перед `run:qa` | passed |
| `QG-S11-D6-07` | Traceability синхронизирована (`issue_map`, `delivery_plan`, sprint/epic docs, history bundle, requirements traceability) | passed |
| `QG-S11-D6-08` | Scope этапа ограничен markdown-only изменениями и без runtime/code edits | passed |

## Definition of Ready / Done для перехода в `run:dev`

### Definition of Ready (`run:dev` launch)
- [x] Design package Day5 (`#454`) подтверждён как source of truth.
- [x] Execution anchor `#458` создан без trigger-лейбла.
- [x] Зафиксированы sequencing-waves между schema foundation, domain semantics, worker continuation, transport bridge, Telegram adapter contour и observability evidence.
- [x] Rollout order `migrations -> control-plane -> worker -> api-gateway -> Telegram adapter contour -> observability/evidence gate` выражен явно.
- [x] Continuity-цепочка `#361 -> #447 -> #448 -> #452 -> #454 -> #456 -> #458` повторно зафиксирована как обязательная.

### Definition of Done (`run:plan` stage)
- [x] Выпущен plan-эпик Day6 с execution package, quality-gates и DoR/DoD.
- [x] Создана handover issue `#458` без trigger-лейбла.
- [x] Обновлены `delivery_plan`, sprint/epic каталоги, `issue_map`, `requirements_traceability` и history-пакет Sprint S11.
- [x] Зафиксированы blockers, risks и owner decisions для следующего stage.

## Self-check (common checklist)
- Проверен scope изменений: только markdown-документы (`*.md`), без кода, YAML/JSON, Dockerfile и shell-скриптов.
- Проверена консистентность stage-policy: `run:dev` остаётся единственным кодовым этапом, trigger-лейбл на issue `#458` автоматически не выставлялся.
- Проверена traceability-синхронизация для нового day6-эпика и issue refs `#456` / `#458`.
- Новые внешние зависимости и версии на этапе Day6 не выбирались; dependency catalog не менялся.
- Секреты, токены и environment values в артефактах не публиковались.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | `BLK-S11-D6-01` | Owner review/approval plan package по Issue `#456` требуется до старта `run:dev` | resolved |
| blocker | `BLK-S11-D6-02` | Trigger-лейбл `run:dev` на issue `#458` должен поставить Owner после review sequencing-waves и quality-gates | open |
| risk | `RSK-S11-D6-01` | Если Wave 1 разрастётся beyond additive schema foundation, ownership drift по `control-plane` заблокирует все последующие waves | monitoring |
| risk | `RSK-S11-D6-02` | Старт Worker/edge/adapter waves до стабилизации Wave 2 приведёт к drift между callback classification и delivery semantics | monitoring |
| risk | `RSK-S11-D6-03` | Если `api-gateway` или adapter contour начнут выводить смысл из raw Telegram payloads, platform-owned interaction semantics будут размыты | monitoring |
| risk | `RSK-S11-D6-04` | Отставание Wave 6 по observability/fallback evidence заблокирует `run:qa`, даже если локально notify/callback path уже работает | monitoring |
| owner-decision | `OD-S11-D6-01` | Реализация идёт только по issue `#458` и по последовательным waves `S11-E01 -> S11-E02 -> S11-E03 -> S11-E04 -> S11-E05 -> S11-E06`; параллельный execution anchor не создаётся | accepted |
| owner-decision | `OD-S11-D6-02` | `run:dev` trigger на issue `#458` выставляется только Owner и только после review/approval plan package Issue `#456` | accepted |
| owner-decision | `OD-S11-D6-03` | Handover в `run:qa` допускается только после acceptance evidence Wave 6 и подтверждённого fallback/manual-action readiness | accepted |
| owner-decision | `OD-S11-D6-04` | Voice/STT, reminders, richer conversation threads, multi-chat routing policy и дополнительные каналы остаются отдельным follow-up контуром вне core Sprint S11 | accepted |

## Tooling validation
- Context7 `/websites/cli_github` использован для актуальной верификации неинтерактивного GitHub CLI flow:
  - `gh issue create`
  - `gh issue edit`
  - `gh pr create`
  - `gh pr edit`
  - `gh pr view`
- Синтаксис команд дополнительно сверен локально:
  - `gh issue create --help`
  - `gh issue edit --help`
  - `gh pr create --help`
  - `gh pr edit --help`
  - `gh pr view --help`
- Новые внешние библиотеки на этапе Day6 не выбирались.

## Acceptance Criteria (Issue #456)
- [x] Подготовлен execution package по потокам schema/control-plane, domain/use-case, worker continuation, thin-edge transport, Telegram adapter contour и observability/evidence gate.
- [x] Для каждого потока зафиксированы scope, зависимости, rollout order и expected success evidence.
- [x] Guardrails сохранены явно: platform-owned semantics, separation from approval flow, typed transport boundary, dependency gate на Sprint S10 и owner-managed review gate.
- [x] Подготовлена follow-up issue `#458` для `run:dev`, где явно повторено continuity-требование продолжить цепочку без разрывов.

## Handover в `run:dev`
- Следующий operational stage: `run:dev`.
- Execution anchor для запуска: issue `#458`.
- Wave order для `#458`: `S11-E01 -> S11-E02 -> S11-E03 -> S11-E04 -> S11-E05 -> S11-E06`.
- Для `run:dev` обязательны:
  - PR с проверками и evidence;
  - синхронное обновление traceability документов;
  - переход в `state:in-review` на PR и Issue после завершения итерации;
  - сохранение issue-цепочки `#361 -> #447 -> #448 -> #452 -> #454 -> #456 -> #458` без разрывов.
