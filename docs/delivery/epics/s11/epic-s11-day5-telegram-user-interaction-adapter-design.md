---
doc_id: EPC-CK8S-S11-D5-TELEGRAM-ADAPTER
type: epic
title: "Epic S11 Day 5: Design для Telegram-адаптера взаимодействия с пользователем (Issues #454/#456)"
status: completed
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-454-design-epic"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# Epic S11 Day 5: Design для Telegram-адаптера взаимодействия с пользователем (Issues #454/#456)

## TL;DR
- Подготовлен полный Day5 design package Sprint S11 для Telegram-адаптера: `design_doc`, `api_contract`, `data_model`, `migrations_policy`.
- Зафиксированы typed contracts для Telegram outbound delivery, inbound normalized callbacks, opaque callback handles, provider message refs, operator visibility и async continuation `edit -> follow-up -> manual fallback`.
- Сохранены platform-owned semantics, separation from approval flow, dependency gate на Sprint S10 interaction foundation и rollout order `migrations -> control-plane -> worker -> api-gateway -> Telegram adapter contour`.
- Создана follow-up issue `#456` для stage `run:plan` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#361` (`docs/delivery/epics/s11/epic-s11-day1-telegram-user-interaction-adapter-intake.md`).
- Vision baseline: `#447` (`docs/delivery/epics/s11/epic-s11-day2-telegram-user-interaction-adapter-vision.md`).
- PRD baseline: `#448` (`docs/delivery/epics/s11/epic-s11-day3-telegram-user-interaction-adapter-prd.md`, `docs/delivery/epics/s11/prd-s11-day3-telegram-user-interaction-adapter.md`).
- Architecture baseline: `#452` (`docs/delivery/epics/s11/epic-s11-day4-telegram-user-interaction-adapter-arch.md` + architecture package).
- Текущий этап: `run:design` в Issue `#454`.
- Scope этапа: только markdown-изменения.

## Design package
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/README.md`
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/design_doc.md`
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/api_contract.md`
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/data_model.md`
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/migrations_policy.md`
- `docs/delivery/traceability/s11_telegram_user_interaction_adapter_history.md`

## Ключевые design-решения
- Callback handle strategy:
  - inline buttons and free-text session path use opaque versioned handles up to `48` ASCII chars;
  - only `sha256` handle hashes are stored in DB;
  - Telegram `callback_data` never carries business semantics.
- Callback auth:
  - adapter -> platform uses interaction-scoped bearer token with TTL `response_deadline_at + 24h grace`;
  - raw Telegram secret-token verification stays outside core platform inside adapter contour.
- Continuation policy:
  - immediate `answerCallbackQuery` stays in adapter contour;
  - business-visible continuation remains async in `worker` as `edit_in_place_first -> follow_up_notify -> manual_fallback_required`.
- Data model:
  - S11 extends S10 interaction foundation with `interaction_channel_bindings`, `interaction_callback_handles`, Telegram evidence fields and operator visibility state;
  - `control-plane` remains the only schema owner.
- Rollout:
  - S10 interaction foundation is a hard prerequisite;
  - S11 rollout can expose notify-only path before decision callbacks and free-text.

## Context7 и внешняя верификация
- Через Context7 подтверждён актуальный Telegram Go SDK baseline:
  - `/mymmrac/telego`
- `go list -m -json github.com/mymmrac/telego@latest` на `2026-03-14` подтвердил latest stable `v1.7.0`.
- Через official Telegram Bot API docs, просмотренные `2026-03-14`, подтверждены design constraints:
  - webhook и `getUpdates` взаимоисключающи;
  - updates хранятся до 24 часов;
  - `setWebhook.secret_token` поддерживает header `X-Telegram-Bot-Api-Secret-Token`;
  - `callback_data` ограничен `1-64 bytes`;
  - `answerCallbackQuery` обязателен для callback UX.

## Acceptance Criteria (Issue #454)
- [x] Подготовлен design package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`).
- [x] Зафиксированы typed delivery/callback contracts, callback handle/token rules и continuation policy.
- [x] Определены schema ownership, migration order, rollback constraints и operator visibility model.
- [x] Сохранены Day4 ownership boundaries, channel-neutral semantics и separation from approval flow.
- [x] Подготовлена follow-up issue `#456` для stage `run:plan`.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S11-D5-01` Contract completeness | Есть `design_doc + api_contract + data_model + migrations_policy` | passed |
| `QG-S11-D5-02` Payload integrity | Callback handle/token strategy укладывается в Telegram limits и не раскрывает semantics | passed |
| `QG-S11-D5-03` Boundary integrity | `control-plane`/`worker`/`api-gateway`/adapter ownership split сохранён | passed |
| `QG-S11-D5-04` Rollout discipline | Зафиксированы S10 prerequisite, additive migrations и continuation rollback constraints | passed |
| `QG-S11-D5-05` Stage continuity | Создана issue `#456` на `run:plan` без trigger-лейбла | passed |

## Handover в `run:plan`
- Следующий этап: `run:plan`.
- Follow-up issue: `#456`.
- Trigger-лейбл на новую issue ставит Owner после review design package.
- На plan-stage обязательно:
  - декомпозировать execution waves по S10 prerequisite, schema, domain, worker continuation, edge transport и adapter rollout;
  - зафиксировать quality gates, DoR/DoD, owner dependencies и acceptance evidence;
  - продолжить issue-цепочку `plan -> dev` без разрывов.
