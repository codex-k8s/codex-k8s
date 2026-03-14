---
doc_id: SPR-CK8S-0011
type: sprint-plan
title: "Sprint S11: Telegram-адаптер взаимодействия с пользователем и первый внешний канал доставки (Issue #361)"
status: completed
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456, 458]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-361-intake"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# Sprint S11: Telegram-адаптер взаимодействия с пользователем и первый внешний канал доставки (Issue #361)

## TL;DR
- Sprint S11 открывает отдельный последовательный product stream для Telegram-адаптера поверх platform-side interaction contract, который формируется в Sprint S10.
- Issue `#361` фиксирует intake baseline: Telegram рассматривается как первый реальный внешний канал доставки/ответа пользователя, но не может стартовать параллельно core stream из Issue `#360`.
- Через Context7 по `/mymmrac/telego` и `go list -m -json github.com/mymmrac/telego@latest` подтверждено, что `github.com/mymmrac/telego v1.7.0` покрывает webhook mode, inline keyboards и callback query handling; библиотека внесена в каталог зависимостей как planned baseline, но не заменяет product/domain contract.
- Intake-пакет ограничивает MVP scope Telegram-канала сценариями `user.notify`, `user.decision.request`, inline callbacks и optional free-text reply, а voice/STT, advanced reminders и richer conversation flows оставляет за пределами core wave.
- Day3 PRD stage выполнен в Issue `#448`: зафиксированы user stories, FR/AC/NFR, expected evidence, callback/webhook guardrails и создана follow-up issue `#452` для `run:arch`; initial continuity issue `#444` остаётся только historical handover artifact.
- Day4 architecture stage выполнен в Issue `#452`: зафиксированы service boundaries, webhook/auth boundary, callback correlation lifecycle, ADR/alternatives и создана follow-up issue `#454` для `run:design`.
- Day5 design stage выполнен в Issue `#454`: выпущен implementation-ready package по typed delivery/callback contracts, callback handles, data model, migrations/rollout policy и создана follow-up issue `#456` для `run:plan`.
- Day6 plan stage выполнен в Issue `#456`: выпущен execution package с sequencing-waves, quality-gates, DoR/DoD и создана follow-up issue `#458` как единый execution anchor для `run:dev`.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan` для инициативы Telegram-адаптера как первого channel-specific stream.
- Формализация продуктовой модели для:
  - доставки `user.notify` в Telegram;
  - доставки `user.decision.request` с 2-5 inline options;
  - приёма callback-ответов и optional free-text reply;
  - базовой webhook/callback security, correlation, idempotency и operability рамки;
  - последовательной зависимости от platform-core interaction contract из Sprint S10.
- Создание последовательных follow-up issue без автоматической постановки `run:*`-лейблов.

### Out of scope
- Кодовая реализация не входит в doc-stage контур Sprint S11 и выполняется только через owner-managed issue `#458`.
- Попытка использовать Telegram как shortcut вместо platform-core contracts Sprint S10.
- Voice/STT, advanced reminders, richer conversation threads, multi-chat routing policy и дополнительные каналы в рамках core Sprint S11.
- Преждевременная фиксация schema/migration/runtime-topology решений до `run:arch` и `run:design`.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Обязательная эскалация:
  - `vision` обязателен, потому что появляется первый channel-specific user-facing experience с отдельными KPI и UX guardrails;
  - `arch` обязателен, потому что scope почти наверняка затрагивает новый adapter contour, callback ingress, security/correlation discipline и операционные границы.
- Целевая continuity-цепочка:
  `#361 (intake) -> #447 (vision) -> #448 (prd) -> #452 (arch) -> #454 (design) -> #456 (plan) -> #458 (dev) -> qa -> release -> postdeploy -> ops`.

## Readiness gate от Sprint S10
- Active `run:prd` stage в Issue `#448` разрешён только после того, как Issue `#389` остаётся закрытой и продолжает ссылаться на design package Issue `#387` как на effective baseline typed interaction contract.
- Проверяемый S10 baseline:
  - `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/data_model.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`.
- По состоянию на `2026-03-14` prerequisite выполнен: Issue `#387` закрыта, Issue `#389` закрыта.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#361`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#447`) | Mission, persona outcomes, KPI/guardrails, MVP/Post-MVP границы | `pm` | Зафиксирован vision baseline и создана continuity issue `#448` для `run:prd` |
| PRD (`#448`) | User stories, FR/AC/NFR, evidence expectations и Telegram-specific edge cases | `pm` + `sa` | Подтверждён PRD package и создана issue `#452` для `run:arch` |
| Architecture (`#452`) | Service boundaries, adapter ownership, callback security/correlation lifecycle | `sa` | Подтверждены архитектурные границы и создана issue `#454` для `run:design` |
| Design (`#454`) | API/data/webhook/runtime contracts и rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue `#456` для `run:plan` |
| Plan (`#456`) | Delivery waves, quality-gates, execution anchor, DoR/DoD | `em` + `km` | Сформирован execution package и создана issue `#458` для owner-managed handover в `run:dev` |

## Guardrails спринта
- Sprint S11 остаётся строго последовательным относительно Sprint S10: Telegram не может задавать core semantics для interaction-domain, а active PRD stage `#448` и follow-up architecture stage `#452` не должны двигаться дальше, если prerequisite из Issue `#389`/`#387` перестаёт быть истинным.
- Telegram adapter должен использовать typed platform interaction contract, а не копировать 1-в-1 поведение reference repositories.
- Базовый MVP ограничен `notify -> decision request -> callback/free-text`; richer conversation UX и voice/STT остаются follow-up scope.
- Inline buttons, callback handling и webhook path считаются обязательным baseline, но они не должны приводить к смешению callback transport и platform-owned domain semantics.
- Telegram callback path должен оставаться UX-safe: callback acknowledgement after button press является обязательным ожиданием продукта, а webhook path должен поддерживать secret-token authenticity expectations.
- Channel-specific UX может оптимизировать delivery experience, но не должен ломать audit trail, correlation discipline и wait-state policy, зафиксированные на platform side.

## Handover
- Документный контур Sprint S11 `intake -> vision -> prd -> arch -> design -> plan` завершён и зафиксирован.
- Plan package:
  - `docs/delivery/sprints/s11/sprint_s11_telegram_user_interaction_adapter.md`;
  - `docs/delivery/epics/s11/epic_s11.md`;
  - `docs/delivery/epics/s11/epic-s11-day6-telegram-user-interaction-adapter-plan.md`;
  - `docs/delivery/traceability/s11_telegram_user_interaction_adapter_history.md`;
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/design_doc.md`;
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/api_contract.md`;
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/data_model.md`;
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/migrations_policy.md`.
- Initial continuity issue `#444` сохранена только как historical handover artifact от intake-stage и 2026-03-14 закрыта как `state:superseded`; vision stage был выполнен в Issue `#447`.
- Следующий stage: `run:dev` в Issue `#458`.
- Проверяемый prerequisite для Issue `#456`: закрытая Issue `#389` с актуальным S10 design package Issue `#387` как baseline typed interaction contract.
- На `2026-03-14` prerequisite уже выполнен и не требует дополнительного parallel launch относительно Sprint S10.
- Входные артефакты от platform-core stream:
  - `docs/delivery/sprints/s10/sprint_s10_mcp_user_interactions.md`;
  - `docs/delivery/epics/s10/epic-s10-day6-mcp-user-interactions-plan.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`.
- Trigger-лейбл для Issue `#458` не ставится автоматически и остаётся owner-managed переходом после завершения doc-flow Sprint S11.
