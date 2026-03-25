---
doc_id: SPR-CK8S-0017
type: sprint-plan
title: "Sprint S17: Unified long-lived user interaction waits and owner feedback inbox (Issue #541)"
status: in-review
owner_role: PM
created_at: 2026-03-20
updated_at: 2026-03-20
related_issues: [360, 361, 458, 473, 532, 540, 541, 554]
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
- Continuity issue `#554` создана для stage `run:vision`; дальнейшие stage issues создаются последовательно после owner review.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan` для unified owner feedback loop.
- Формализация продуктовой модели для:
  - same-session continuation как primary happy-path;
  - recovery-only snapshot resume;
  - 24h long human-wait policy;
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
  `#541 (intake) -> #554 (vision) -> PRD issue -> architecture issue -> design issue -> plan issue -> dev execution waves -> qa -> release -> postdeploy -> ops`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#541`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#554`) | Mission, north star, persona outcomes, KPI/guardrails, wave boundaries | `pm` | Зафиксирован vision baseline и создана issue для `run:prd` |
| PRD (TBD) | User stories, FR/AC/NFR, expected evidence и edge cases | `pm` + `sa` | Подтверждён PRD package и создана issue для `run:arch` |
| Architecture (TBD) | Execution model, ownership split, lifetime policy, continuation semantics | `sa` | Подтверждены архитектурные границы и создана issue для `run:design` |
| Design (TBD) | API/data/UI/runtime contracts и rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue для `run:plan` |
| Plan (TBD) | Delivery waves, execution issues, DoR/DoD, quality-gates | `em` + `km` | Сформирован execution package и owner-managed handover в `run:dev` |

## Guardrails спринта
- Same live pod / same `codex` session остаётся primary happy-path и не заменяется detached resume-run без нового owner-решения.
- Persisted session snapshot используется только как recovery fallback при потере live runtime.
- Long human wait не меньше 24 часов должен отражаться одновременно в interaction TTL, wait-state semantics, pod lifetime expectations и timeout policy.
- Delivery lifecycle обязан разделять `delivery pending` и `waiting for user response`.
- Telegram inbox и staff-console fallback обязаны использовать общий persisted backend contract; канал не может становиться owner of semantics.
- `run:self-improve` остаётся вне owner-facing human-wait contract.
- До `run:plan` Sprint S17 остаётся markdown-only и не создаёт code/runtime diff.

## Handover
- Текущий stage in-review: `run:intake` в Issue `#541`.
- Следующий stage: `run:vision` через issue `#554`.
- До завершения следующего stage нельзя потерять следующие Day1 decisions:
  - same live session как primary continuation model;
  - snapshot-resume как recovery-only fallback;
  - long human-wait target `>=24h`;
  - delivery-before-wait lifecycle;
  - Telegram pending inbox + staff-console fallback;
  - deterministic text/voice binding;
  - self-improve exclusion.
- Trigger-лейбл для issue `#554` не ставится автоматически и остаётся owner-managed переходом после review intake package.
