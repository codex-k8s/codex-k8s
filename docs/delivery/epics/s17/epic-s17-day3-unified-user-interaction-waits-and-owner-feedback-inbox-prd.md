---
doc_id: EPC-CK8S-S17-D3-OWNER-FEEDBACK-WAITS
type: epic
title: "Epic S17 Day 3: PRD для unified long-lived user interaction waits и owner feedback inbox (Issues #557/#559)"
status: in-review
owner_role: PM
created_at: 2026-03-25
updated_at: 2026-03-25
related_issues: [360, 361, 458, 532, 540, 541, 554, 557, 559]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-25-issue-557-prd"
---

# Epic S17 Day 3: PRD для unified long-lived user interaction waits и owner feedback inbox (Issues #557/#559)

## TL;DR
- Подготовлен PRD-пакет Sprint S17 для unified owner feedback loop: `epic-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox-prd.md` и `prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`.
- Зафиксированы user stories, FR/AC/NFR, scenario matrix, edge cases и expected evidence для owner inbox, same-session continuation, max timeout/TTL baseline built-in `kodex` MCP wait path, lifecycle transparency, deterministic response binding и recovery-only fallback.
- Принято продуктовое решение: same live pod / same `codex` session остаётся primary happy-path, snapshot-resume используется только как recovery fallback, Telegram pending inbox и staff-console fallback работают поверх одного persisted backend contract, а overdue/manual-fallback visibility остаётся blocking requirement.
- Дополнительные каналы, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как равноправный happy-path остаются later-wave scope.
- Создана follow-up issue `#559` для stage `run:arch` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#541` (`docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md`).
- Vision baseline: `#554` (`docs/delivery/epics/s17/epic-s17-day2-unified-user-interaction-waits-and-owner-feedback-inbox-vision.md`).
- Текущий этап: `run:prd` в Issue `#557`.
- Следующий этап: `run:arch` в Issue `#559`.
- Входной product contract:
  - `docs/delivery/sprints/s10/sprint_s10_mcp_user_interactions.md`;
  - `docs/delivery/sprints/s11/sprint_s11_telegram_user_interaction_adapter.md`;
  - `docs/architecture/mcp_approval_and_audit_flow.md`;
  - `docs/architecture/prompt_templates_policy.md`.

## Scope
### In scope
- Формализация user stories, FR/AC/NFR, scenario matrix и expected evidence для unified owner feedback loop.
- Приоритизация волн `core same-session owner feedback -> transparency/recovery evidence -> deferred conversation expansion`.
- Фиксация product guardrails для same-session continuation, max timeout/TTL baseline built-in `kodex` MCP wait path, recovery-only snapshot-resume, dual-channel inbox, deterministic text/voice/callback binding и overdue/manual-fallback transparency.
- Явный handover в `run:arch` с перечнем продуктовых решений, которые нельзя потерять.
- Синхронизация traceability (`issue_map`, `delivery_plan`, sprint/epic docs, history package, indexes).

### Out of scope
- Кодовая реализация, storage/schema/runtime topology и transport/API lock-in до `run:arch` / `run:design`.
- Возврат detached resume-run как primary happy-path.
- Попытка сделать Telegram-first UX источником platform semantics.
- Additional channels, reminders, attachments, multi-party routing и richer conversation UX в рамках core MVP.

## PRD package
- `docs/delivery/epics/s17/epic-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox-prd.md`
- `docs/delivery/epics/s17/prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`
- `docs/delivery/traceability/s17_unified_user_interaction_waits_and_owner_feedback_inbox_history.md`

## Wave priorities

| Wave | Приоритет | Scope | Exit signal |
|---|---|---|---|
| Wave 1 | `P0` | Pending inbox через Telegram + staff-console fallback, same-session continuation, delivery-before-wait lifecycle, max timeout/TTL baseline built-in `kodex` MCP wait path | Owner получает pending request, понимает, что агент реально ждёт, и ответ приводит к continuation без GitHub-comment detour |
| Wave 2 | `P0` | Deterministic text/voice/callback binding, overdue / expired / manual-fallback transparency, recovery-only snapshot-resume classification и expected evidence | Negative/runtime edge cases explainable, а fallback path остаётся явно классифицированным и не скрывает product truth |
| Wave 3 | `P1` (deferred) | Additional channels, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как отдельный product stream | Stream двигается только после owner-approved architecture/design без reopening core same-session contract |

## Acceptance criteria (Issue #557)
- [x] Подготовлен PRD-артефакт unified owner feedback loop и синхронизирован в traceability-документах.
- [x] Для core flows зафиксированы user stories, FR/AC/NFR, scenario matrix, edge cases и expected evidence.
- [x] Явно сохранены locked baselines: same live pod / same `codex` session как primary happy-path, max timeout/TTL built-in `kodex` MCP wait path не ниже owner wait window, snapshot-resume только как recovery fallback, lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`, Telegram pending inbox + staff-console fallback, deterministic text/voice binding и `run:self-improve` exclusion.
- [x] Wave priorities сформулированы без смешения core MVP и deferred conversation/channel expansion scope.
- [x] Создана follow-up issue `#559` для stage `run:arch` без trigger-лейбла.

## Quality gates

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S17-D3-01 PRD completeness | User stories, FR/AC/NFR, scenario matrix и expected evidence покрывают scope Day3 | passed |
| QG-S17-D3-02 Locked baseline preserved | Same-session happy-path, max timeout/TTL baseline, recovery-only fallback, dual-channel inbox и lifecycle transparency сохранены без reopening | passed |
| QG-S17-D3-03 Deferred-scope discipline | Additional channels, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run не смешаны с core MVP | passed |
| QG-S17-D3-04 Stage continuity | Создана issue `#559` для `run:arch` без trigger-лейбла и с continuity-требованием `arch -> design -> plan -> dev` | passed |
| QG-S17-D3-05 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#559`.
- Trigger-лейбл `run:arch` на issue `#559` ставит Owner после review PRD-пакета.
- Обязательные выходы архитектурного этапа:
  - service boundaries и ownership matrix для `control-plane`, `worker`, `agent-runner`, `api-gateway`, `staff web-console` и `telegram-interaction-adapter`;
  - alternatives/ADR по live wait lifetime, persisted request truth, response binding/correlation, recovery fallback и visibility path;
  - фиксация, как сохраняются same-session happy-path, max timeout/TTL baseline, recovery-only snapshot-resume, delivery-before-wait lifecycle и dual-channel semantics;
  - отдельная issue на `run:design` без trigger-лейбла после завершения `run:arch` с явным continuity-требованием продолжить цепочку `design -> plan -> dev`.

## Открытые риски и допущения

| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | `RSK-557-01` | Same-session contract может деградировать в operational shortcut через shorter timeout или hidden resume | open |
| risk | `RSK-557-02` | Dual-channel inbox может потерять parity и превратиться в два расходящихся источника истины | open |
| risk | `RSK-557-03` | Deterministic text/voice/callback binding и overdue/manual-fallback visibility могут остаться недоопределёнными до architecture stage | open |
| assumption | `ASM-557-01` | Telegram pending inbox и staff-console fallback достаточно для core MVP без расширения на дополнительные каналы | accepted |
| assumption | `ASM-557-02` | Большинство pilot-сценариев можно покрыть same-session continuation, оставив snapshot-resume исключительным recovery path | accepted |
