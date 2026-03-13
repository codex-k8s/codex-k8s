---
doc_id: EPC-CK8S-S10-D3-MCP-INTERACTIONS
type: epic
title: "Epic S10 Day 3: PRD для built-in MCP user interactions (Issues #383/#385)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-383-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic S10 Day 3: PRD для built-in MCP user interactions (Issues #383/#385)

## TL;DR
- Подготовлен PRD-пакет Sprint S10 для built-in MCP user interactions: `epic-s10-day3-mcp-user-interactions-prd.md` и `prd-s10-day3-mcp-user-interactions.md`.
- Зафиксированы user stories, FR/AC/NFR, edge cases, expected evidence и wave priorities для `user.notify`, `user.decision.request`, wait-state discipline, typed response semantics и adapter-neutral interaction domain.
- Принято продуктовое решение: interaction flow остаётся отдельным доменом относительно approval flow, а Telegram/adapters, richer conversations, voice/STT и delivery preferences не блокируют core MVP.
- Создана follow-up issue `#385` для stage `run:arch` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#360` (`docs/delivery/epics/s10/epic-s10-day1-mcp-user-interactions-intake.md`).
- Vision baseline: `#378` (`docs/delivery/epics/s10/epic-s10-day2-mcp-user-interactions-vision.md`).
- Текущий этап: `run:prd` в Issue `#383`.
- Следующий этап: `run:arch` в Issue `#385`.

## Scope
### In scope
- Формализация user stories, FR/AC/NFR и edge cases для built-in MCP user interactions.
- Приоритизация волн `core MVP -> platform evidence -> deferred adapter streams`.
- Фиксация product guardrails для `user.notify`, `user.decision.request`, typed response semantics, wait-state lifecycle и adapter-neutral contract.
- Явный handover в `run:arch` с перечнем продуктовых решений, которые нельзя потерять.
- Синхронизация traceability (`issue_map`, `delivery_plan`, sprint/epic docs, history package).

### Out of scope
- Кодовая реализация, storage/schema decisions и transport/runtime lock-in.
- Telegram-first UX, voice/STT, richer multi-turn conversation threads и advanced delivery policies.
- Попытка использовать approval flow как shortcut для user interactions.
- Отдельный runtime server block вместо расширения built-in `codex_k8s`.

## PRD package
- `docs/delivery/epics/s10/epic-s10-day3-mcp-user-interactions-prd.md`
- `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`
- `docs/delivery/traceability/s10_mcp_user_interactions_history.md`

## Wave priorities

| Wave | Приоритет | Scope | Exit signal |
|---|---|---|---|
| Wave 1 | `P0` | `user.notify` как actionable completion/next-step path, `user.decision.request` с typed option/free-text response, separation from approval flow | Агент и пользователь могут закрыть result/decision сценарии без GitHub-comment detour и без approval-only shortcuts |
| Wave 2 | `P0` | Wait-state lifecycle, callback validity, retries/idempotency expectations, audit/correlation evidence и adapter-ready semantics | Late/duplicate/invalid responses не ломают state machine, а platform evidence достаточно для acceptance и architecture handover |
| Wave 3 | `P1` (deferred) | Telegram/adapters, delivery preferences, reminders, richer conversation threads, voice/STT | Stream входит в roadmap только после подтверждения channel-neutral core contract на `run:arch` и `run:design` |

## Acceptance criteria (Issue #383)
- [x] Подготовлен PRD-артефакт built-in MCP user interactions и синхронизирован в traceability-документах.
- [x] Для core flows зафиксированы user stories, FR/AC/NFR, edge cases и expected evidence.
- [x] Wave priorities сформулированы без смешения core MVP и adapter-specific follow-up streams.
- [x] Сохранены неподвижные ограничения инициативы: built-in server `codex_k8s`, separation from approval flow, non-blocking `user.notify`, wait-state только для `user.decision.request`, platform-owned audit/correlation semantics.
- [x] Создана follow-up issue `#385` для stage `run:arch` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S10-D3-01 PRD completeness | User stories, FR/AC/NFR, edge cases и expected evidence покрывают scope Day3 | passed |
| QG-S10-D3-02 Wave discipline | Core MVP и deferred adapter streams разделены по приоритетам и exit signals | passed |
| QG-S10-D3-03 Guardrails preserved | Approval flow separation, wait-state discipline и built-in server baseline сохранены | passed |
| QG-S10-D3-04 Stage continuity | Создана issue `#385` для `run:arch` без trigger-лейбла | passed |
| QG-S10-D3-05 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#385`.
- Trigger-лейбл `run:arch` на issue `#385` ставит Owner после review PRD-пакета.
- Обязательные выходы архитектурного этапа:
  - service boundaries и ownership для request orchestration, callback ingestion, wait-state transitions, audit/correlation и future adapters;
  - alternatives/ADR по lifecycle ownership, callback path, persisted state и adapter isolation без потери product contract;
  - фиксация, как сохраняются separation from approval flow, non-blocking `user.notify` и typed response semantics;
  - отдельная issue на `run:design` без trigger-лейбла после завершения `run:arch`.

## Открытые риски и допущения
| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | RSK-383-01 | Инициатива может расползтись в Telegram-first или approval-first решение вместо platform-owned interaction domain | open |
| risk | RSK-383-02 | Typed response model окажется слишком свободной и даст неоднозначную интерпретацию callback data | open |
| risk | RSK-383-03 | Ownership wait-state, retries и correlation останется размытым между сервисами до `run:arch` | open |
| assumption | ASM-383-01 | Existing built-in server `codex_k8s` достаточно расширяем для новых interaction tools без отдельного runtime server block | accepted |
| assumption | ASM-383-02 | Пользовательская ценность достигается быстрее через typed options/free-text path, чем через GitHub comments и ручные follow-up | accepted |
