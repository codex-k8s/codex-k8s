---
doc_id: TRH-CK8S-S17-0001
type: traceability-history
title: "Sprint S17 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-20
updated_at: 2026-04-01
related_issues: [360, 361, 458, 473, 532, 540, 541, 554, 557, 559, 568, 575, 582]
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

## Актуализация по Issue #554 (`run:vision`, 2026-03-25)
- Подготовлен vision package:
  - `docs/delivery/epics/s17/epic-s17-day2-unified-user-interaction-waits-and-owner-feedback-inbox-vision.md`;
  - обновлены `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md` и `docs/delivery/epics/s17/epic_s17.md`.
- Зафиксированы:
  - unified owner feedback loop как platform capability, где owner отвечает в Telegram или staff-console, а агент продолжает ту же задачу без GitHub-comment detour и без channel drift;
  - mission, north star, persona outcomes, KPI/guardrails и wave boundaries для owner/product lead path, same-session runtime path и staff/operator fallback path;
  - locked baseline Day1: same live pod / same `codex` session как primary happy-path, snapshot-resume только как recovery fallback, long human-wait target `>=24h`, delivery-before-wait lifecycle, Telegram pending inbox, staff-console fallback, deterministic text/voice binding и `run:self-improve` exclusion;
  - owner revise-замечание дополнительно зафиксировано в vision baseline: для built-in `kodex` MCP wait path effective timeout/TTL должен быть максимальным и не ниже owner wait window, чтобы happy-path был реальным live wait на tool response, а не resume с подложенным tool result;
  - additional channels, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как равноправный happy-path выведены в later-wave scope.
- Через `gh issue create` создана follow-up issue `#557` для stage `run:prd` с continuity-требованием сохранить цепочку `prd -> arch -> design -> plan -> dev`.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue view 554 --json number,title,body,url`, `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope и не требовал runtime-debug.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: vision stage добавляет product framing, guardrails и historical delta, не создавая новых канонических FR/NFR.

## Актуализация по Issue #557 (`run:prd`, 2026-03-25)
- Подготовлен PRD package:
  - `docs/delivery/epics/s17/epic-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox-prd.md`;
  - `docs/delivery/epics/s17/prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`;
  - обновлены `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md` и `docs/delivery/epics/s17/epic_s17.md`.
- Зафиксированы:
  - unified owner feedback loop как product contract для owner inbox, same-session continuation, delivery-before-wait lifecycle и dual-channel semantics;
  - user stories, FR/AC/NFR, scenario matrix и expected evidence для owner/product lead path, live runtime path и staff/operator fallback;
  - blocking baseline: same live pod / same `codex` session как primary happy-path, max timeout/TTL built-in `kodex` MCP wait path не ниже owner wait window, snapshot-resume только как recovery fallback, long human-wait target `>=24h`, lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`, Telegram pending inbox, staff-console fallback, deterministic text/voice binding и `run:self-improve` exclusion;
  - explicit later-wave boundary: дополнительные каналы, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как равноправный happy-path остаются deferred scope;
  - expected product evidence для same-session continuity, recovery classification, overdue/manual-fallback transparency и GitHub-comment fallback как degraded path.
- Через `gh issue create` создана follow-up issue `#559` для stage `run:arch` с continuity-требованием сохранить цепочку `arch -> design -> plan -> dev`.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue view 557 --json number,title,body,url`, `gh issue view 559 --json number,title,body,url`, `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope и не требовал runtime-debug.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: PRD stage закрепляет sprint-specific product contract и historical delta без изменения канонического requirements baseline.

## Актуализация по Issue #559 (`run:arch`, 2026-03-26)
- Подготовлен architecture package:
  - `docs/delivery/epics/s17/epic-s17-day4-unified-user-interaction-waits-and-owner-feedback-inbox-arch.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/README.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/architecture.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/c4_context.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/c4_container.md`;
  - `docs/architecture/adr/ADR-0017-unified-owner-feedback-loop-live-wait-primary-platform-owned-continuation.md`;
  - `docs/architecture/alternatives/ALT-0009-unified-owner-feedback-loop-live-wait-and-channel-ownership.md`;
  - обновлены `docs/architecture/README.md`, `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md`, `docs/delivery/epics/s17/epic_s17.md` и `docs/delivery/delivery_plan.md`.
- Зафиксированы:
  - `control-plane` как единственный owner feedback request truth, accepted-response winner, wait/deadline policy и continuation classification;
  - `worker` как owner dispatch/retry/reconcile и runtime lease keepalive для long-lived waits, при этом `agent-runner` остаётся owner only for live same-session execution and recovery snapshot capture;
  - `api-gateway`, `staff web-console` и `telegram-interaction-adapter` как thin surfaces вокруг одного persisted backend contract без права переопределять lifecycle semantics;
  - primary happy-path = same live pod / same `codex` session; effective max timeout/TTL built-in `kodex` MCP wait path не ниже owner wait window; snapshot-resume допускается только как recovery fallback;
  - canonical visibility model для `overdue`, `expired`, `manual-fallback` и `recovery-resume`, чтобы degraded paths не оставались hidden operator-only detail.
- Через `gh issue create` создана follow-up issue `#568` для stage `run:design` с continuity-требованием сохранить цепочку `design -> plan -> dev`.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue view 559 --json number,title,body,url`, `gh issue view 568 --json number,title,body,url`, `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope и не требовал runtime-debug.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: architecture stage закрепляет service boundaries, trade-offs и historical delta без изменения канонического requirements baseline.

## Актуализация по Issue #568 (`run:design`, 2026-03-27)
- Подготовлен design package:
  - `docs/delivery/epics/s17/epic-s17-day5-unified-user-interaction-waits-and-owner-feedback-inbox-design.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/README.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/design_doc.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/api_contract.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/data_model.md`;
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/migrations_policy.md`;
  - обновлены `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md`, `docs/delivery/epics/s17/epic_s17.md`, `docs/delivery/delivery_plan.md` и `docs/delivery/issue_map.md`.
- Зафиксированы:
  - built-in wait entrypoint остаётся на `user.decision.request`, а control tool `owner.feedback.request` не переиспользуется для ordinary owner responses;
  - persisted owner-feedback truth materializes как additive overlay поверх Sprint S10/S11 interaction foundation: `interaction_requests` extensions + `owner_feedback_wait_links` + `owner_feedback_channel_projections` + `owner_feedback_response_bindings`;
  - Telegram callback/free-text/voice и staff-console fallback responses сходятся в один response binding registry и одну winner-selection policy;
  - staff-console оформлен как projection + typed response surface, а recovery resume остаётся explicit degraded path с отдельным `continuation_path`.
- Через `gh issue create` создана follow-up issue `#575` для stage `run:plan` с continuity-требованием сохранить цепочку `plan -> dev`.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue view 568 --json number,title,body,url`, `gh issue view 575 --json number,title,body,url`, `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope и не требовал runtime-debug.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: design stage закрепляет implementation-ready contracts и historical delta без изменения канонического requirements baseline.

## Актуализация по Issue #575 (`run:plan`, 2026-04-01)
- Подготовлен plan package:
  - `docs/delivery/epics/s17/epic-s17-day6-unified-user-interaction-waits-and-owner-feedback-inbox-plan.md`;
  - обновлены `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md`, `docs/delivery/epics/s17/epic_s17.md`, `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, `docs/delivery/epics/README.md`, `docs/delivery/sprints/README.md` и `docs/delivery/traceability/README.md`.
- Зафиксированы:
  - execution package `S17-E01..S17-E07` для schema ownership, domain/use-case, worker visibility, `api-gateway`, `telegram-interaction-adapter`, `staff web-console` и observability/evidence gate;
  - prerequisite gate на закрытых Sprint S10/S11 foundation issues `#391..#395` и `#458`;
  - single execution anchor `#582` для `run:dev`, обязательный rollout order и owner-managed trigger policy;
  - quality-gates, DoR/DoD, blockers, risks и owner decisions без пересмотра Day1-Day5 baseline.
- Через `gh issue create` создана follow-up issue `#582` для stage `run:dev` с continuity-требованием сохранить цепочку `#541 -> #554 -> #557 -> #559 -> #568 -> #575 -> #582`.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue view 391 --json number,title,state,url`, `gh issue view 392 --json number,title,state,url`, `gh issue view 393 --json number,title,state,url`, `gh issue view 394 --json number,title,state,url`, `gh issue view 395 --json number,title,state,url`, `gh issue view 458 --json number,title,state,url`, `gh issue view 575 --json number,title,body,url`, `gh issue view 582 --json number,title,body,url`, `gh issue create --help`, `gh issue edit --help`, `gh pr create --help`, `gh pr edit --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope и не требовал runtime-debug.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: plan stage закрепляет execution governance, handover и historical delta, не создавая новых канонических требований.
