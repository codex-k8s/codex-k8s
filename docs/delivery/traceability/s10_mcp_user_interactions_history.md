---
doc_id: TRH-CK8S-S10-0001
type: traceability-history
title: "Sprint S10 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387, 389, 391, 392, 393, 394, 395, 402]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-traceability-s10-history"
---

# Sprint S10 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S10.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #360 (`run:intake`, 2026-03-12)
- Intake зафиксировал built-in MCP user interactions как отдельную product initiative поверх существующего built-in server `codex_k8s`.
- В качестве baseline зафиксированы:
  - MVP tools `user.notify` и `user.decision.request`;
  - channel-neutral interaction-domain;
  - раздельные semantics для approval flow и user interaction flow;
  - wait-state только для response-required сценариев;
  - Telegram как отдельный последовательный follow-up stream.
- Создана continuity issue `#378` для stage `run:vision`.
- Root FR/NFR matrix не менялась: intake stage не обновляет канонический requirements baseline, а фиксирует problem/scope/handover для нового delivery stream.

## Актуализация по Issue #378 (`run:vision`, 2026-03-12)
- Подготовлен vision package:
  - `docs/delivery/epics/s10/epic-s10-day2-mcp-user-interactions-vision.md`.
- Зафиксированы:
  - mission и north star для built-in MCP user interactions как отдельной channel-neutral capability платформы;
  - persona outcomes для owner/product lead, end user и platform operator;
  - KPI/guardrails для actionable notifications, decision turnaround, fallback-to-comments, separation from approval flow и correlation correctness;
  - явное разделение core MVP и deferred streams: Telegram/adapters, voice/STT, richer threads и advanced delivery policies не блокируют core baseline.
- Для continuity создана follow-up issue `#383` (`run:prd`) без trigger-лейбла.
- Попытка использовать Context7 для GitHub CLI manual завершилась ошибкой `Monthly quota exceeded`; неинтерактивный issue/PR flow дополнительно сверен локально по `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась, потому что vision stage уточняет mission, KPI и scope boundaries, но не меняет канонический requirements baseline.

## Актуализация по Issue #383 (`run:prd`, 2026-03-12)
- Подготовлен PRD package:
  - `docs/delivery/epics/s10/epic-s10-day3-mcp-user-interactions-prd.md`;
  - `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`.
- Зафиксированы:
  - user stories, FR/AC/NFR и wave priorities для `user.notify`, `user.decision.request`, typed response semantics и adapter-neutral contract;
  - explicit edge cases для stale/duplicate/invalid responses, fallback-to-comments и separation from approval flow;
  - handover decisions, которые нельзя потерять на `run:arch`: built-in `codex_k8s`, non-blocking `user.notify`, wait-state только для `user.decision.request`, platform-owned audit/correlation/retry semantics и deferred scope для Telegram/adapters.
- Для continuity создана follow-up issue `#385` (`run:arch`) без trigger-лейбла.
- Попытка использовать Context7 для GitHub CLI manual снова завершилась ошибкой `Monthly quota exceeded`; для non-interactive GitHub flow использованы локальные `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: PRD stage уточняет product contract и delivery evidence, а в root-матрице синхронизирована только связь по issue/traceability governance.

## Актуализация по Issue #385 (`run:arch`, 2026-03-12)
- Подготовлен architecture package:
  - `docs/delivery/epics/s10/epic-s10-day4-mcp-user-interactions-arch.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/c4_context.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/c4_container.md`;
  - `docs/architecture/adr/ADR-0012-built-in-mcp-user-interactions-control-plane-owned-lifecycle.md`;
  - `docs/architecture/alternatives/ALT-0004-built-in-mcp-user-interactions-lifecycle-boundaries.md`.
- Зафиксированы:
  - ownership split между `control-plane`, `worker`, `api-gateway` и future adapters;
  - отдельный interaction-domain без reuse approval-specific semantics как source-of-truth;
  - lifecycle `tool call -> dispatch -> callback -> resume` с platform-owned retry/idempotency/expiry/audit expectations.
- Для continuity создана follow-up issue `#387` (`run:design`) без trigger-лейбла.
- Попытка использовать Context7 для Mermaid/C4 documentation завершилась ошибкой `Monthly quota exceeded`; для пакета использованы существующие Mermaid/C4 conventions репозитория.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: architecture stage фиксирует service boundaries и handover в design, а не меняет канонический requirements baseline.

## Актуализация по Issue #387 (`run:design`, 2026-03-12)
- Подготовлен design package:
  - `docs/delivery/epics/s10/epic-s10-day5-mcp-user-interactions-design.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/data_model.md`;
  - `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`.
- Зафиксированы:
  - typed contracts для `user.notify`, `user.decision.request`, outbound adapter envelope и inbound callback family;
  - отдельный persisted interaction-domain: aggregate, delivery attempts, callback evidence, response records;
  - wait-state taxonomy с сохранением coarse runtime state `waiting_mcp`, но с отдельным `wait_reason=interaction_response` и typed wait linkage;
  - resume contract через deterministic `interaction_resume_payload`, использующий existing `agent_sessions` snapshot path без reuse approval tables.
- Для continuity создана follow-up issue `#389` (`run:plan`) без trigger-лейбла.
- Попытка использовать Context7 для `kin-openapi` и `goose` завершилась ошибкой `Monthly quota exceeded`; новые внешние зависимости на этапе `run:design` не добавлялись.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: design stage конкретизирует API/data/runtime contracts и migration policy, но не меняет канонический product baseline.

## Актуализация по Issue #389 (`run:plan`, 2026-03-13)
- Подготовлен plan package:
  - `docs/delivery/epics/s10/epic-s10-day6-mcp-user-interactions-plan.md`.
- Зафиксированы:
  - execution waves `#391..#395` для `control-plane` foundation, worker dispatch/retry/expiry, contract-first callback ingress, deterministic resume path в `agent-runner` и observability/evidence gate;
  - sequencing `#391 -> #392 -> #393/#394 -> #395` с сохранением rollout order `migrations -> control-plane -> worker -> api-gateway` и отдельным resume gate;
  - DoR/DoD, blockers/risks/owner decisions и запрет на auto-trigger labels для implementation issues;
  - channel-specific adapters, Telegram, reminders и voice/STT оставлены вне core Sprint S10 execution package.
- Для continuity созданы follow-up issues `#391`, `#392`, `#393`, `#394`, `#395` без trigger-лейблов.
- Попытка использовать Context7 для GitHub CLI manual завершилась ошибкой `Monthly quota exceeded`; для non-interactive GitHub flow использованы локальные `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: plan stage фиксирует execution backlog, quality gates и handover в `run:dev`, а не меняет канонический requirements baseline.

## Актуализация по Issue #391 (`run:dev`, 2026-03-13)
- Реализован foundation package для stream `S10-E01` в `services/internal/control-plane`:
  - миграция `20260313120000_day29_mcp_user_interactions_foundation.sql` добавила таблицы `interaction_requests`, `interaction_delivery_attempts`, `interaction_callback_events`, `interaction_response_records`, а также typed wait columns в `agent_runs` с backfill из legacy wait-state данных;
  - доменные typed models/repositories добавлены для interaction aggregate, delivery attempts, callback evidence и resume payload;
  - built-in MCP orchestration получила foundation для `user.notify` и `user.decision.request`, callback classification/resume helpers и audit events;
  - staff run SQL projections переведены на `agent_runs.wait_reason` с fallback к coarse `agent_sessions.wait_state`, чтобы сохранить совместимость текущих wait filters до следующих волн.
- Rollout boundary сохранён:
  - новые tools зарегистрированы в MCP catalog/transport, но не добавлены в allow-list policy, поэтому остаются скрытыми до delivery waves `#392`, `#393`, `#394`;
  - callback ingress / OpenAPI / gRPC transport на этой волне не открывались.
- Выполнены проверки:
  - `go test ./services/internal/control-plane/...`
  - `git diff --check`
- Для residual `dupl-go` blockers вне scope создан follow-up issue `#402`.
- Попытка использовать Context7 для актуальной документации/verification завершилась ошибкой `Monthly quota exceeded`; новых внешних зависимостей в issue `#391` не добавлялось.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: issue `#391` реализует approved foundation wave из execution package, не меняя продуктовый baseline.
