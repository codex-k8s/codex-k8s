---
doc_id: EPC-CK8S-S11-D4-TELEGRAM-ADAPTER
type: epic
title: "Epic S11 Day 4: Architecture для Telegram-адаптера взаимодействия с пользователем (Issues #452/#454)"
status: completed
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-452-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# Epic S11 Day 4: Architecture для Telegram-адаптера взаимодействия с пользователем (Issues #452/#454)

## TL;DR
- Подготовлен architecture package Sprint S11 для Telegram-адаптера: architecture decomposition, C4 overlays, ADR-0014 и alternatives по ownership, webhook/auth boundary, callback correlation и fallback policy.
- Зафиксирован ownership split для built-in interaction semantics, outbound delivery/retries, raw Telegram webhook/auth, normalized callback ingress, post-callback edit/follow-up behaviour и operator visibility.
- Подготовлен handover в `run:design` без premature transport/schema lock-in.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#361` (`docs/delivery/epics/s11/epic-s11-day1-telegram-user-interaction-adapter-intake.md`).
- Vision baseline: `#447` (`docs/delivery/epics/s11/epic-s11-day2-telegram-user-interaction-adapter-vision.md`).
- PRD baseline: `#448` (`docs/delivery/epics/s11/epic-s11-day3-telegram-user-interaction-adapter-prd.md`, `docs/delivery/epics/s11/prd-s11-day3-telegram-user-interaction-adapter.md`).
- Текущий этап: `run:arch` в Issue `#452`.
- Scope этапа: только markdown-изменения.

## Architecture package
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/README.md`
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/architecture.md`
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/c4_context.md`
- `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/c4_container.md`
- `docs/architecture/adr/ADR-0014-telegram-user-interaction-adapter-platform-owned-lifecycle.md`
- `docs/architecture/alternatives/ALT-0006-telegram-user-interaction-adapter-boundaries.md`
- `docs/delivery/traceability/s11_telegram_user_interaction_adapter_history.md`

## Ключевые решения Stage
- `control-plane` остаётся владельцем interaction semantics, correlation, replay/expiry classification, wait-state transitions и operator-visible outcome.
- `worker` закреплён за outbound delivery, retries, expiry scans и post-callback UX continuation; edit-vs-follow-up behaviour вынесен из callback ingress path в async platform-owned side effect.
- Raw Telegram webhooks, secret-token verification и callback query acknowledgement остаются во внешнем Telegram adapter contour; `api-gateway` принимает только normalized callbacks с platform-issued auth.
- Callback payload direction зафиксирован как opaque/server-side lookup strategy, а не как Telegram-first semantic payload model.

## Context7 и внешний baseline
- Context7 использован для проверки Telegram Go SDK baseline:
  - `/mymmrac/telego`.
- Context7 использован для проверки Telegram Bot API constraints:
  - `/websites/core_telegram_bots_api`.
- Official Telegram Bot API docs (`Getting updates`, `setWebhook`, callback behaviour) просмотрены 2026-03-14 и использованы как внешний baseline для webhook/auth/callback guardrails.

## Acceptance Criteria (Issue #452)
- [x] Подготовлен architecture package с service boundaries, ownership, C4 overlays, ADR и alternatives для Telegram-адаптера как первого внешнего channel-specific stream.
- [x] Для core flows определены owner-сервисы и границы ответственности: interaction semantics, delivery/retry, raw webhook/auth, normalized callback ingress, fallback policy и operator visibility.
- [x] Зафиксированы architecture-level trade-offs по callback payload strategy, webhook authenticity boundary и edit-vs-follow-up policy без premature transport/storage lock-in.
- [x] Обновлены delivery/traceability документы и package indexes.
- [x] Подготовлена follow-up issue `#454` для stage `run:design` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S11-D4-01` Architecture completeness | Есть package `architecture + C4 + ADR + alternatives` | passed |
| `QG-S11-D4-02` Boundary integrity | Ownership за `control-plane` / `worker` / `api-gateway` / Telegram adapter contour зафиксирован явно | passed |
| `QG-S11-D4-03` Transport isolation | Raw Telegram webhook/auth отделён от platform callback ingress | passed |
| `QG-S11-D4-04` Semantic neutrality | Callback payload direction и outcome semantics не Telegram-first | passed |
| `QG-S11-D4-05` Stage continuity | Подготовлена issue `#454` на `run:design` без trigger-лейбла | passed |

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#454`.
- Trigger-лейбл на новую issue ставит Owner после review architecture package.
- На design-stage обязательно:
  - выпустить `design_doc`, `api_contract`, `data_model`, `migrations_policy`;
  - определить exact callback handle/token rules, typed delivery/callback DTO и persistence model для provider refs;
  - зафиксировать rollout/rollback notes и продолжить issue-цепочку `design -> plan -> dev`.
