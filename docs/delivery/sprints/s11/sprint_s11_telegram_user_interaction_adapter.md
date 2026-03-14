---
doc_id: SPR-CK8S-0011
type: sprint-plan
title: "Sprint S11: Telegram-адаптер взаимодействия с пользователем и первый внешний канал доставки (Issue #361)"
status: in-review
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-361-intake"
---

# Sprint S11: Telegram-адаптер взаимодействия с пользователем и первый внешний канал доставки (Issue #361)

## TL;DR
- Sprint S11 открывает отдельный последовательный product stream для Telegram-адаптера поверх platform-side interaction contract, который формируется в Sprint S10.
- Issue `#361` фиксирует intake baseline: Telegram рассматривается как первый реальный внешний канал доставки/ответа пользователя, но не может стартовать параллельно core stream из Issue `#360`.
- Через Context7 по `/mymmrac/telego` и `go list -m -json github.com/mymmrac/telego@latest` подтверждено, что `github.com/mymmrac/telego v1.7.0` покрывает webhook mode, inline keyboards и callback query handling; библиотека внесена в каталог зависимостей как planned baseline, но не заменяет product/domain contract.
- Intake-пакет ограничивает MVP scope Telegram-канала сценариями `user.notify`, `user.decision.request`, inline callbacks и optional free-text reply, а voice/STT, advanced reminders и richer conversation flows оставляет за пределами core wave.
- Active Day2 vision stage выполняется в Issue `#447`, initial continuity issue `#444` сохраняется как исторический handover от intake, а follow-up issue `#448` подготовлена для `run:prd`.

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
- Кодовая реализация до завершения и утверждения `run:plan`.
- Попытка использовать Telegram как shortcut вместо platform-core contracts Sprint S10.
- Voice/STT, advanced reminders, richer conversation threads, multi-chat routing policy и дополнительные каналы в рамках core Sprint S11.
- Преждевременная фиксация schema/migration/runtime-topology решений до `run:arch` и `run:design`.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Обязательная эскалация:
  - `vision` обязателен, потому что появляется первый channel-specific user-facing experience с отдельными KPI и UX guardrails;
  - `arch` обязателен, потому что scope почти наверняка затрагивает новый adapter contour, callback ingress, security/correlation discipline и операционные границы.
- Целевая continuity-цепочка:
  `#361 (intake) -> #447 (vision) -> #448 (prd) -> arch -> design -> plan -> dev -> qa -> release -> postdeploy -> ops`.

## Readiness gate от Sprint S10
- Active `run:vision` stage в Issue `#447` разрешён только после того, как Issue `#389` остаётся закрытой и продолжает ссылаться на design package Issue `#387` как на effective baseline typed interaction contract.
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
| PRD (`#448`) | User stories, FR/AC/NFR, evidence expectations и Telegram-specific edge cases | `pm` + `sa` | Подтверждён PRD package и создана issue для `run:arch` |
| Architecture | Service boundaries, adapter ownership, callback security/correlation lifecycle | `sa` | Подтверждены архитектурные границы и создана issue для `run:design` |
| Design | API/data/webhook/runtime contracts и rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue для `run:plan` |
| Plan | Delivery waves, quality-gates, execution issues, DoR/DoD | `em` + `km` | Сформирован execution package и owner-managed handover в `run:dev` |

## Guardrails спринта
- Sprint S11 остаётся строго последовательным относительно Sprint S10: Telegram не может задавать core semantics для interaction-domain, а active stage `#447` не должен двигаться дальше, если prerequisite из Issue `#389`/`#387` перестаёт быть истинным.
- Telegram adapter должен использовать typed platform interaction contract, а не копировать 1-в-1 поведение reference repositories.
- Базовый MVP ограничен `notify -> decision request -> callback/free-text`; richer conversation UX и voice/STT остаются follow-up scope.
- Inline buttons, callback handling и webhook path считаются обязательным baseline, но они не должны приводить к смешению callback transport и platform-owned domain semantics.
- Channel-specific UX может оптимизировать delivery experience, но не должен ломать audit trail, correlation discipline и wait-state policy, зафиксированные на platform side.

## Handover
- Текущий stage in-review: `run:vision` в Issue `#447`.
- Vision package:
  - `docs/delivery/sprints/s11/sprint_s11_telegram_user_interaction_adapter.md`;
  - `docs/delivery/epics/s11/epic_s11.md`;
  - `docs/delivery/epics/s11/epic-s11-day1-telegram-user-interaction-adapter-intake.md`.
- `docs/delivery/epics/s11/epic-s11-day2-telegram-user-interaction-adapter-vision.md`.
- Initial continuity issue `#444` остаётся historical handover artifact от intake-stage; active vision stage выполняется в Issue `#447`.
- Следующий stage: `run:prd` в Issue `#448`.
- Проверяемый prerequisite для Issue `#447`: закрытая Issue `#389` с актуальным S10 design package Issue `#387` как baseline typed interaction contract.
- На `2026-03-14` prerequisite уже выполнен и не требует дополнительного parallel launch относительно Sprint S10.
- Входные артефакты от platform-core stream:
  - `docs/delivery/sprints/s10/sprint_s10_mcp_user_interactions.md`;
  - `docs/delivery/epics/s10/epic-s10-day6-mcp-user-interactions-plan.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`.
- Trigger-лейбл для Issue `#448` не ставится автоматически и остаётся owner-managed переходом после review vision package.
