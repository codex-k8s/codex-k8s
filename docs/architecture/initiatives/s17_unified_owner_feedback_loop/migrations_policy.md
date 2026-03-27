---
doc_id: MIG-S17-CK8S-0001
type: migrations-policy
title: "Sprint S17 Day 5 — Migrations policy for unified owner feedback loop (Issue #568)"
status: in-review
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [541, 554, 557, 559, 568, 575]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-568-migrations-policy"
---

# DB Migrations Policy: Sprint S17 unified owner feedback loop

## TL;DR
- Подход: additive expand-enable over Sprint S10/S11 interaction foundation.
- Инструменты миграций: `goose` in schema owner `services/internal/control-plane`.
- Владелец схемы и миграций: `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Политика rollback: disable owner-feedback write exposure and keep additive evidence; accepted responses and recovery classifications are not rolled back by destructive schema operations.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- S17 does not introduce a new DB owner.
- `api-gateway`, `worker`, `telegram-interaction-adapter` and `web-console` do not get separate migration paths.

## Пререквизиты Sprint S10/S11
- S17 migrations may start only after:
  - Sprint S10 interaction foundation is deployed:
    - generic interaction tables,
    - typed wait linkage in `agent_runs`,
    - resume payload lookup path;
  - Sprint S11 Telegram extension is deployed:
    - `interaction_channel_bindings`,
    - `interaction_callback_handles`,
    - Telegram provider refs / callback evidence path.
- If either prerequisite is missing:
  - S17 rollout is blocked;
  - no partial owner-feedback schema branch is allowed.

## Потоки и миграционная обязательность
| Stream | Нужна миграция | Политика |
|---|---|---|
| `interaction_requests` owner-feedback fields | да | additive columns |
| `owner_feedback_wait_links` | да | новая таблица |
| `owner_feedback_channel_projections` | да | новая таблица |
| `owner_feedback_response_bindings` | да | новая таблица + unique hash index |
| `interaction_response_records` owner-feedback fields | да | additive columns |
| `agent_runs` continuation hints | да | additive columns |
| `agent_sessions` | нет | reuse existing heartbeat/snapshot fields |
| `flow_events` payload vocabulary | нет | event schema evolves in code, no table rewrite |

## Принципы
- Expand first:
  - S17 overlay tables and columns land before any new owner-feedback write path is enabled.
- Live-wait discipline:
  - no migration may shorten effective wait timeout/TTL below owner wait window.
- Surface parity:
  - staff projection schema is additive and cannot become second request truth.
- Recovery remains explicit:
  - recovery state and continuation path are additive classification fields, not hidden transport metadata.
- Writes last:
  - callback/staff response admission is enabled only after schema, indexes and owner services are ready.

## Процесс миграции (run:dev target)
1. Verify prerequisites:
   - confirm Sprint S10/S11 foundation migrations already applied;
   - confirm no unknown `wait_reason` or legacy interaction table names remain.
2. Expand schema:
   - add owner-feedback columns to `interaction_requests`;
   - create `owner_feedback_wait_links`;
   - create `owner_feedback_channel_projections`;
   - create `owner_feedback_response_bindings`;
   - add owner-feedback fields to `interaction_response_records`;
   - add continuation hints to `agent_runs`.
3. Index hardening:
   - wait deadline / recovery state index;
   - projection list index;
   - unique binding-hash index;
   - partial unique effective-response index if not already present.
4. Enable owner writes:
   - rollout `control-plane` with owner-feedback aggregate, bindings and wait-link management.
5. Enable reconcile and visibility:
   - rollout `worker` with delivery accepted / overdue / expired / manual-fallback / recovery transitions.
6. Enable transport admission:
   - rollout `api-gateway` staff endpoints and callback bridge;
   - rollout `telegram-interaction-adapter` voice/text/callback normalization.
7. Enable UI response path:
   - rollout `web-console` read-only projection first;
   - then enable typed fallback response submission.

## Как выполняются миграции при деплое
- Mandatory order:
  1. stateful dependencies ready
  2. Sprint S10/S11 prerequisite confirmed
  3. S17 migration job
  4. `control-plane`
  5. `worker`
  6. `api-gateway`
  7. `telegram-interaction-adapter`
  8. `web-console`
- Concurrency control:
  - single migration runner under `goose` advisory lock.
- Failure policy:
  - if migration or prerequisite verification fails, owner-feedback traffic stays disabled and UI remains read-only or hidden.

## Политика backfill
- No historical backfill into S17 owner-feedback overlays is required:
  - existing S10/S11 interactions stay `interaction_family=generic`;
  - new owner-feedback rows are created only after S17 write path is enabled.
- Partial rollout restart safety:
  - if request aggregate exists but projections/bindings are missing after crash, `control-plane` reconstructs them idempotently from canonical request row.
- `agent_sessions` snapshots are not rewritten during migration.

## Политика rollback
- Safe rollback before S17 traffic:
  - keep additive schema;
  - disable owner-feedback write exposure and new UI/actions.
- Limited rollback after Telegram/staff traffic:
  - stop new owner-feedback requests;
  - keep request truth, projection evidence and response bindings;
  - keep staff UI in read-only visibility mode if response submission is unstable.
- Continuation-specific rollback:
  - disable Telegram voice normalization first while keeping option/text response path;
  - if staff response path is unstable, disable write endpoint but keep read model for manual diagnosis.

## Что нельзя безопасно откатить
- Accepted owner responses and their resume payloads.
- Recovery classification that already distinguished `continuation_live` vs `recovery_resume`.
- Visibility evidence for `overdue`, `expired` and `manual_fallback`.
- Delivered Telegram messages and already published staff projections referenced by audit evidence.

## Проверки
### Pre-migration checks
- Sprint S10/S11 prerequisite schema is deployed and healthy.
- No conflicting custom tables/columns exist with S17 names.
- Existing `agent_sessions` wait-state and snapshot fields are populated on active runs.
- Planned owner wait window policy is not shorter than effective built-in MCP timeout/TTL baseline.

### Post-migration verification
- `user.decision.request` can create an owner-feedback row with `interaction_family=owner_feedback`.
- One request produces one open `owner_feedback_wait_links` row.
- Telegram callback and staff response paths classify duplicates/stale replies deterministically.
- `manual_fallback` projection remains queryable without changing request truth.
- Recovery resume persists `continuation_path=recovery_resume` instead of rewriting terminal history.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): none.
- Migration impact (`run:dev`): moderate, additive schema over S10/S11 foundation with new tables, columns and indexes.

## Operational notes
- Staff projection read-only mode can be enabled before staff response submission.
- Telegram voice normalization can be dark-launched after text/callback paths are stable.
- If live-session evidence is unreliable, rollout must stop before enabling recovery resume automation.

## Апрув
- request_id: owner-2026-03-27-issue-568-migrations-policy
- Решение: pending
- Комментарий: Ожидается review additive rollout policy и mixed-version safety relative to Sprint S10/S11.
