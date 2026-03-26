---
doc_id: EPC-CK8S-S17-D4-OWNER-FEEDBACK-ARCH
type: epic
title: "Epic S17 Day 4: Architecture для unified long-lived user interaction waits и owner feedback inbox (Issues #559/#568)"
status: in-review
owner_role: SA
created_at: 2026-03-26
updated_at: 2026-03-26
related_issues: [541, 554, 557, 559, 568]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-26-issue-559-arch"
---

# Epic S17 Day 4: Architecture для unified long-lived user interaction waits и owner feedback inbox (Issues #559/#568)

## TL;DR
- Подготовлен architecture package Sprint S17 для unified owner feedback loop: `architecture.md`, `c4_context.md`, `c4_container.md`, `ADR-0017` и `ALT-0009`.
- Зафиксированы service boundaries между `control-plane`, `worker`, `agent-runner`, `api-gateway`, `staff web-console` и `telegram-interaction-adapter`, а также live wait lifetime policy, persisted truth ownership и deterministic binding/correlation.
- Принято архитектурное решение: same live pod / same `codex` session остаётся primary happy-path, snapshot-resume используется только как recovery fallback, а Telegram и staff-console остаются thin surfaces поверх одного persisted backend truth.
- Создана follow-up issue `#568` для stage `run:design` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#541` (`docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md`).
- Vision baseline: `#554` (`docs/delivery/epics/s17/epic-s17-day2-unified-user-interaction-waits-and-owner-feedback-inbox-vision.md`).
- PRD baseline: `#557` (`docs/delivery/epics/s17/epic-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox-prd.md`, `docs/delivery/epics/s17/prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`).
- Текущий этап: `run:arch` в Issue `#559`.
- Следующий этап: `run:design` в Issue `#568`.
- Входной архитектурный baseline:
  - `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/architecture.md`
  - `docs/architecture/agent_runtime_rbac.md`
  - `docs/architecture/mcp_approval_and_audit_flow.md`

## Scope
### In scope
- Фиксация service boundaries и ownership split для unified owner feedback loop.
- Архитектурные решения по live wait lifetime, max timeout/TTL baseline, recovery-only snapshot-resume, persisted request truth и dual-channel parity.
- Alternatives/ADR по execution model и channel ownership.
- Явный handover в `run:design` с continuity-требованием сохранить цепочку `design -> plan -> dev`.
- Синхронизация traceability (`issue_map`, `delivery_plan`, sprint/epic docs, history package, architecture index).

### Out of scope
- Точные DTO, schema columns, migrations, OpenAPI/grpc details и UI layouts.
- Кодовая реализация, runtime-manifests, shell/scripts и любые non-markdown изменения.
- Additional channels, reminders/escalations, attachments, generalized conversation platform и detached resume-run как равноправный happy-path.

## Architecture package
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/README.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/architecture.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/c4_context.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/c4_container.md`
- `docs/architecture/adr/ADR-0017-unified-owner-feedback-loop-live-wait-primary-platform-owned-continuation.md`
- `docs/architecture/alternatives/ALT-0009-unified-owner-feedback-loop-live-wait-and-channel-ownership.md`

## Key decisions
- `control-plane` владеет feedback request aggregate, lifecycle truth, accepted-response winner, continuation classification и channel parity.
- `worker` владеет async delivery/retry/reconcile, runtime lease keepalive и visibility transitions для `overdue / expired / manual-fallback`.
- `agent-runner` удерживает live same-session execution и recovery snapshot capture, но не владеет persisted request truth.
- `api-gateway`, `staff web-console` и `telegram-interaction-adapter` остаются thin surfaces вокруг одного persisted backend contract.
- Detached resume-run не допускается как normal primary path; successful snapshot restore классифицируется только как recovery fallback.

## Acceptance criteria (Issue #559)
- [x] Подготовлен architecture package Sprint S17 и синхронизирован в traceability-документах.
- [x] Зафиксированы service boundaries, ownership matrix, live wait lifetime policy, persisted request truth и wait/continuation semantics.
- [x] Сохранены blocking baselines: same-session happy-path, max timeout/TTL built-in `codex_k8s` MCP wait path не ниже owner wait window, recovery-only snapshot-resume, delivery-before-wait lifecycle, dual-channel inbox и `run:self-improve` exclusion.
- [x] Design-level schema/API/UI решения оставлены следующему stage и не подменяют architecture package.
- [x] Создана follow-up issue `#568` для stage `run:design` без trigger-лейбла.

## Quality gates

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S17-D4-01 Boundary integrity | Все service boundaries и owners зафиксированы без premature new-service split | passed |
| QG-S17-D4-02 Same-session baseline | Live wait primary path, timeout/TTL baseline и recovery-only snapshot-resume сохранены без reopening | passed |
| QG-S17-D4-03 Channel parity | Telegram и staff-console оставлены thin surfaces поверх одного persisted truth | passed |
| QG-S17-D4-04 Visibility discipline | Overdue / expired / manual-fallback states оформлены как platform-visible outcomes | passed |
| QG-S17-D4-05 Stage continuity | Создана issue `#568` для `run:design` без trigger-лейбла и с continuity `design -> plan -> dev` | passed |
| QG-S17-D4-06 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#568`.
- Trigger-лейбл `run:design` на issue `#568` ставит Owner после review architecture package.
- Обязательные выходы design stage:
  - `design_doc.md`, `api_contract.md`, `data_model.md`, `migrations_policy.md` в `docs/architecture/initiatives/s17_unified_owner_feedback_loop/`;
  - typed contracts для built-in wait path, Telegram callbacks и staff-console fallback actions;
  - data model для request truth, projections, deterministic binding, stale/duplicate handling и visibility states;
  - rollout/rollback notes для long-lived waits и continuity issue для `run:plan`.

## Открытые риски и допущения

| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | `RSK-559-01` | Long-lived wait policy может разойтись между MCP timeout/TTL, runtime lease и UI visibility, если design stage не удержит один logical wait window | open |
| risk | `RSK-559-02` | Staff-console fallback может превратиться во второй source of truth, если read model и action model будут спроектированы ad-hoc | open |
| risk | `RSK-559-03` | Deterministic binding для text/voice/callback replies может остаться недоопределённой до design stage | open |
| assumption | `ASM-559-01` | Существующие bounded contexts `control-plane` + `worker` + thin surfaces достаточно сильны для MVP без отдельного owner-feedback coordinator service | accepted |
