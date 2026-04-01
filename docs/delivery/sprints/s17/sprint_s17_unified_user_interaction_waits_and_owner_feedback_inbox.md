---
doc_id: SPR-CK8S-0017
type: sprint-plan
title: "Sprint S17: Unified long-lived user interaction waits and owner feedback inbox (Issue #541)"
status: in-review
owner_role: PM
created_at: 2026-03-20
updated_at: 2026-04-01
related_issues: [360, 361, 458, 473, 532, 540, 541, 554, 557, 559, 568, 575, 582]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-20-issue-541-intake"
---

# Sprint S17: Unified long-lived user interaction waits and owner feedback inbox (Issue #541)

## TL;DR
- Цель спринта: превратить built-in feedback tools и existing Telegram channel path в единый owner feedback loop, где пользователь отвечает в понятном inbox, а агент детерминированно продолжает ту же задачу.
- Intake stage в Issue `#541` уже зафиксировал ключевой baseline: primary happy-path = same live pod / same `codex` session until user response, while snapshot-resume is recovery-only fallback.
- Sprint S17 также фиксирует обязательные guardrails: long human-wait target `>=24h`, delivery-before-wait lifecycle, Telegram pending inbox, staff-console fallback, persisted text/voice binding и self-improve exclusion.
- Vision package в Issue `#554` зафиксировал mission, north star, persona outcomes, KPI/guardrails и wave boundaries для unified owner feedback loop, не переоткрывая Day1 baseline, и явно добавил product guardrail: built-in `codex_k8s` MCP wait path обязан использовать максимальный timeout/TTL не ниже owner wait window.
- PRD package в Issue `#557` формализовал user stories, FR/AC/NFR, scenario matrix, expected evidence и recovery/lifecycle guardrails для owner feedback loop без reopening Day1/Day2 baseline.
- Architecture package в Issue `#559` закрепил service boundaries, live wait lifetime policy, persisted request truth и recovery-only snapshot-resume boundary без переоткрытия Day1-Day3 baseline.
- Design package в Issue `#568` закрепил implementation-ready API/data/runtime contracts, response binding registry, wait-state linkage и rollout policy без пересмотра Day1-Day4 baseline.
- Plan package в Issue `#575` зафиксировал execution waves `S17-E01..S17-E07`, prerequisite gate на Sprint S10/S11 foundation, quality-gates, DoR/DoD и создал handover issue `#582` для `run:dev`.
- Handover issue `#582` создана как единый execution anchor для owner-managed запуска `run:dev`; дальнейшие stage issues создаются последовательно после review.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan` для unified owner feedback loop.
- Формализация продуктовой модели для:
  - same-session continuation как primary happy-path;
  - recovery-only snapshot resume;
  - 24h long human-wait policy;
  - max timeout/TTL для built-in `codex_k8s` MCP wait path, чтобы agent pod реально ждал ответ tool в той же session;
  - delivery-before-wait lifecycle visibility;
  - Telegram pending inbox;
  - staff-console fallback;
  - deterministic text/voice binding;
  - unified continuation semantics для всех run-типов, кроме `run:self-improve`.
- Создание последовательных follow-up issue без `run:*`-лейблов; trigger следующего запуска остаётся owner-managed.

### Out of scope
- Кодовая реализация до завершения и утверждения `run:plan`.
- Редизайн approval flow, который не связан напрямую с human-wait contract.
- Дополнительные каналы помимо Telegram и staff-console fallback.
- Advanced reminders, attachments, multi-party routing и generalized conversation platform.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Обоснование:
  - инициатива меняет platform-wide execution semantics и owner-facing UX для нескольких bounded contexts сразу;
  - нужны обязательные `vision`, `arch` и `design`, чтобы удержать same-session baseline, cost/recovery trade-offs и channel-neutral contract;
  - сокращённые launch profile не удержат cross-service impact и continuity discipline.
- Целевая continuity-цепочка:
  `#541 (intake) -> #554 (vision) -> #557 (prd) -> #559 (arch) -> #568 (design) -> #575 (plan) -> #582 (dev execution anchor) -> qa -> release -> postdeploy -> ops`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#541`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#554`) | Mission, north star, persona outcomes, KPI/guardrails, wave boundaries | `pm` | Зафиксирован vision baseline и создана issue `#557` для `run:prd` |
| PRD (`#557`) | User stories, FR/AC/NFR, scenario matrix, expected evidence и edge cases | `pm` + `sa` | Подтверждён PRD package и создана issue `#559` для `run:arch` |
| Architecture (`#559`) | Execution model, ownership split, lifetime policy, continuation semantics | `sa` | Подтверждены архитектурные границы и создана issue `#568` для `run:design` |
| Design (`#568`) | API/data/UI/runtime contracts и rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue `#575` для `run:plan` |
| Plan (`#575`) | Delivery waves, execution anchor, DoR/DoD, quality-gates | `em` + `km` | Сформирован execution package и создана issue `#582` для owner-managed handover в `run:dev` |
| Dev (`#582`) | Additive schema/domain/runtime/transport/UI implementation, tests/evidence, PR | `dev` | Выполнена dev-итерация, подготовлен PR и собрана candidate evidence для handover в `run:qa` |

## Guardrails спринта
- Same live pod / same `codex` session остаётся primary happy-path и не заменяется detached resume-run без нового owner-решения.
- Persisted session snapshot используется только как recovery fallback при потере live runtime.
- Long human wait не меньше 24 часов должен отражаться одновременно в interaction TTL, wait-state semantics, pod lifetime expectations, timeout policy и max timeout/TTL built-in `codex_k8s` MCP wait path.
- Delivery lifecycle обязан разделять `delivery pending` и `waiting for user response`.
- Telegram inbox и staff-console fallback обязаны использовать общий persisted backend contract; канал не может становиться owner of semantics.
- `run:self-improve` остаётся вне owner-facing human-wait contract.
- До owner-managed запуска `run:dev` в issue `#582` Sprint S17 остаётся markdown-only и не создаёт code/runtime diff.
- `run:dev` выполняется только через единый execution anchor `#582` и сохраняет prerequisite gate на закрытых Sprint S10/S11 issues `#391..#395` и `#458`.

## Handover
- Текущий stage in-review: `run:plan` в Issue `#575`.
- Plan package:
  - `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md`;
  - `docs/delivery/epics/s17/epic_s17.md`;
  - `docs/delivery/epics/s17/epic-s17-day6-unified-user-interaction-waits-and-owner-feedback-inbox-plan.md`;
  - `docs/delivery/traceability/s17_unified_user_interaction_waits_and_owner_feedback_inbox_history.md`;
  - follow-up issue `#582` для `run:dev`.
- Design package:
  - `docs/delivery/epics/s17/epic-s17-day5-unified-user-interaction-waits-and-owner-feedback-inbox-design.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/design_doc.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/api_contract.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/data_model.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/migrations_policy.md`.
- Следующий stage: `run:dev` через issue `#582`.
- Проверяемые foundation prerequisites для `run:dev`: issues `#391`, `#392`, `#393`, `#394`, `#395` и `#458` должны оставаться закрытыми.
- Wave order для issue `#582`: `S17-E01 -> S17-E02 -> S17-E03 -> S17-E04 -> S17-E05 -> S17-E06 -> S17-E07`.
- До завершения следующего stage нельзя потерять следующие Day1/Day2/Day3 decisions:
  - same live session как primary continuation model;
  - max timeout/TTL built-in `codex_k8s` MCP wait path не ниже owner wait window, чтобы happy-path оставался live wait, а не synthetic resume;
  - snapshot-resume как recovery-only fallback;
  - long human-wait target `>=24h`;
  - delivery-before-wait lifecycle;
  - Telegram pending inbox + staff-console fallback;
  - deterministic text/voice binding;
  - visibility для overdue / expired / manual-fallback scenarios;
  - self-improve exclusion.
  - `control-plane` как owner persisted request truth, accepted-response winner и continuation classification;
  - `worker` как owner dispatch/reconcile/lease keepalive, а `agent-runner` только как owner live session и recovery snapshot capture;
  - `api-gateway`, `staff web-console` и `telegram-interaction-adapter` как thin surfaces вокруг одного persisted backend contract.
  - response binding registry и additive overlay tables поверх Sprint S10/S11 interaction foundation.
- Trigger-лейбл для issue `#582` не ставится автоматически и остаётся owner-managed переходом после review plan package.
