---
doc_id: TRH-CK8S-S17-0001
type: traceability-history
title: "Sprint S17 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-20
updated_at: 2026-03-20
related_issues: [360, 361, 458, 473, 532, 540, 541, 554]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-20-traceability-s17-history"
---

# Sprint S17 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S17.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #541 (`run:intake`, 2026-03-20)
- Подготовлен intake package:
  - `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md`;
  - `docs/delivery/epics/s17/epic_s17.md`;
  - `docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md`.
- Зафиксированы:
  - Sprint S17 как отдельная cross-cutting initiative по unified long-lived human-wait contract, deterministic continuation semantics и owner-facing inbox;
  - hybrid execution baseline: same live pod / same `codex` session до user response как happy-path, snapshot-resume только как recovery fallback;
  - long human-wait target `>=24h` и конфликт текущего repo baseline `tool_timeout_sec = 180` с owner-driven wait window;
  - явный lifecycle `created -> delivery pending -> delivery accepted -> waiting for user response -> response received -> continuation resumed`;
  - Telegram pending inbox и staff-console fallback как единый owner-facing contour поверх общего persisted backend contract;
  - persisted text/voice binding и deterministic continuation после inline/text/voice reply;
  - `run:self-improve` как исключение из human-wait contract.
- Через `gh issue create` создана continuity issue `#554` для stage `run:vision`.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope и использует issue-driven production evidence.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: intake stage фиксирует problem/scope/handover и historical delta, а не добавляет новые канонические требования в `docs/product/requirements_machine_driven.md`.
