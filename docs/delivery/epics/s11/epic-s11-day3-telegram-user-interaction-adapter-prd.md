---
doc_id: EPC-CK8S-S11-D3-TELEGRAM-ADAPTER
type: epic
title: "Epic S11 Day 3: PRD для Telegram-адаптера взаимодействия с пользователем (Issues #448/#452)"
status: completed
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-448-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# Epic S11 Day 3: PRD для Telegram-адаптера взаимодействия с пользователем (Issues #448/#452)

## TL;DR
- Подготовлен PRD-пакет Sprint S11 для Telegram-адаптера: `epic-s11-day3-telegram-user-interaction-adapter-prd.md` и `prd-s11-day3-telegram-user-interaction-adapter.md`.
- Зафиксированы user stories, FR/AC/NFR, edge cases, expected evidence и wave priorities для `user.notify`, `user.decision.request`, inline callbacks, optional free-text reply и operability guardrails первого внешнего channel-specific stream.
- Принято продуктовое решение: Telegram остаётся adapter-layer реализацией поверх platform-owned interaction semantics Sprint S10, а approval flow, voice/STT, rich conversation threads, advanced reminders и дополнительные каналы не входят в core MVP.
- Через Context7 подтверждено, что `/mymmrac/telego` покрывает webhook mode, text updates, inline keyboards и callback query handling, а `go list -m -json github.com/mymmrac/telego@latest` на 2026-03-14 подтверждает latest stable `v1.7.0`; reference SDK остаётся implementation baseline, а не product contract.
- Создана follow-up issue `#452` для stage `run:arch` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#361` (`docs/delivery/epics/s11/epic-s11-day1-telegram-user-interaction-adapter-intake.md`).
- Vision baseline: `#447` (`docs/delivery/epics/s11/epic-s11-day2-telegram-user-interaction-adapter-vision.md`).
- Текущий этап: `run:prd` в Issue `#448`.
- Следующий этап: `run:arch` в Issue `#452`.
- Входной product contract Sprint S10:
  - `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/data_model.md`.

## Scope
### In scope
- Формализация user stories, FR/AC/NFR и edge cases для Telegram-адаптера как первого внешнего channel-specific stream.
- Приоритизация волн `core Telegram MVP -> callback safety and evidence -> deferred channel expansion`.
- Фиксация product guardrails для `user.notify`, `user.decision.request`, inline callbacks, optional free-text, typed outcome semantics и separation from approval flow.
- Явный handover в `run:arch` с перечнем product decisions, которые нельзя потерять.
- Синхронизация traceability (`issue_map`, `requirements_traceability`, `delivery_plan`, sprint/epic docs, history package).

### Out of scope
- Кодовая реализация, storage/schema decisions и transport/runtime lock-in.
- Telegram-first redesign core interaction contract Sprint S10.
- Voice/STT, advanced reminders, rich multi-turn conversation threads, multi-chat routing policy и дополнительные каналы.
- Прямое копирование reference repositories как готовой architecture baseline.

## PRD package
- `docs/delivery/epics/s11/epic-s11-day3-telegram-user-interaction-adapter-prd.md`
- `docs/delivery/epics/s11/prd-s11-day3-telegram-user-interaction-adapter.md`
- `docs/delivery/traceability/s11_telegram_user_interaction_adapter_history.md`

## Wave priorities

| Wave | Приоритет | Scope | Exit signal |
|---|---|---|---|
| Wave 1 | `P0` | Telegram delivery для `user.notify`, `user.decision.request`, inline callbacks, optional free-text и non-GitHub response path | Пользователь получает actionable notification или даёт валидный typed ответ в Telegram без обязательного ухода в GitHub comments |
| Wave 2 | `P0` | Callback safety, duplicate/replay/expired handling, webhook authenticity expectations, observability/audit evidence и operator fallback clarity | Late/duplicate/invalid callback scenarios не ломают platform lifecycle, а Telegram UX остаётся adapter-safe и explainable |
| Wave 3 | `P1` (deferred) | Voice/STT, reminders, richer conversation threads, multi-chat routing и дополнительные каналы | Stream входит в roadmap только после подтверждения core architecture и design package без потери channel-neutral semantics |

## Acceptance criteria (Issue #448)
- [x] Подготовлен PRD-артефакт Telegram-адаптера и синхронизирован в traceability-документах.
- [x] Для core flows зафиксированы user stories, FR/AC/NFR, edge cases и expected evidence.
- [x] Wave priorities сформулированы без смешения core MVP и deferred channel-expansion scope.
- [x] Сохранены неподвижные ограничения инициативы: sequencing gate Sprint S10, separation from approval flow, platform-owned semantics, Telegram как adapter-layer stream, а не source of truth.
- [x] Создана follow-up issue `#452` для stage `run:arch` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S11-D3-01 PRD completeness | User stories, FR/AC/NFR, edge cases и expected evidence покрывают scope Day3 | passed |
| QG-S11-D3-02 Callback and operability guardrails | Typed callback semantics, free-text fallback, webhook/callback safety и operator visibility зафиксированы | passed |
| QG-S11-D3-03 Deferred scope discipline | Voice/STT, reminders, rich threads и дополнительные каналы не смешаны с core MVP | passed |
| QG-S11-D3-04 Stage continuity | Создана issue `#452` для `run:arch` без trigger-лейбла и с продолжением цепочки до `run:dev` | passed |
| QG-S11-D3-05 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#452`.
- Trigger-лейбл `run:arch` на issue `#452` ставит Owner после review PRD-пакета.
- Обязательные выходы архитектурного этапа:
  - service boundaries и ownership для callback ingestion, Telegram delivery path, wait-state transitions, audit/correlation и operator visibility;
  - alternatives/ADR по callback payload strategy, webhook security/correlation lifecycle, message update/edit policy и adapter isolation без потери product contract;
  - фиксация, как сохраняются sequencing gate Sprint S10, separation from approval flow и platform-owned typed semantics;
  - отдельная issue на `run:design` без trigger-лейбла после завершения `run:arch`, с повторным continuity-требованием довести цепочку до `run:dev`.

## Открытые риски и допущения
| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | `RSK-448-01` | Инициатива может расползтись в Telegram-first UX вместо adapter-layer канала поверх platform-owned interaction contract | open |
| risk | `RSK-448-02` | Callback/free-text path может стать удобным для пользователя, но хрупким по duplicate/replay/expired сценариям и operator diagnostics | open |
| risk | `RSK-448-03` | Ownership Telegram delivery, callback lifecycle и audit/correlation останется размытым между edge, jobs и domain services до `run:arch` | open |
| assumption | `ASM-448-01` | Notify + decision request + inline callbacks + optional free-text достаточно, чтобы подтвердить ценность первого внешнего канала без richer conversations | accepted |
| assumption | `ASM-448-02` | Telegram может ускорить user decision turnaround без потери channel-neutral semantics Sprint S10 | accepted |
