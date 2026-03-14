---
doc_id: MIG-S9-MISSION-CONTROL-0001
type: migrations-policy
title: "Mission Control Dashboard — DB migrations policy Sprint S9 Day 5"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340, 351, 363]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-351-migrations-policy"
---

# DB Migrations Policy: Mission Control Dashboard

## TL;DR
- Подход: additive expand-warmup-enable, без destructive rewrite существующих платформенных таблиц.
- Владелец схемы/миграций: `services/internal/control-plane`.
- Миграции лежат в `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Rollback ограничен после включения inline write-path, потому что provider side effects нельзя отменить простым schema rollback.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- Все DDL создаются внутри owner service migrations directory:
  - `services/internal/control-plane/cmd/cli/migrations/*.sql`
- Shared DB without owner not allowed; `worker` and `api-gateway` не получают собственные миграции для Mission Control Dashboard.

## Принципы
- Additive first:
  - новые таблицы и индексы создаются без ломки текущего runtime.
- Warmup before exposure:
  - read endpoints и realtime path включаются только после projection warmup/backfill.
- Writes last:
  - inline commands включаются после верификации snapshot freshness и reconcile correctness.
- Voice isolated:
  - `voice_candidates` схема и write-path включаются отдельным rollout step и отдельным feature flag.

## Процесс миграции (план для `run:dev`)
1. Expand:
   - создать таблицы `mission_control_entities`, `mission_control_relations`, `mission_control_timeline_entries`, `mission_control_commands`;
   - в `mission_control_commands` сразу добавить approval-поля (`approval_request_id`, `approval_state`, `approval_requested_at`, `approval_decided_at`) и closed enum со статусом `pending_approval`;
   - опционально создать `mission_control_voice_candidates` в той же волне либо второй additive migration.
2. Index hardening:
   - unique key on entity external identity;
   - dedupe indexes for timeline/provider deliveries and `business_intent_key`;
   - lookup index for approval queue (`project_id`, `approval_state`, `updated_at desc`);
   - query indexes for active-state and timeline sorting.
3. Warmup/backfill:
   - запустить rebuild job под owner-логикой `control-plane` / execution `worker`;
   - собрать initial active set из issue/PR/discussion/run/provider state;
   - materialize timeline projection and relation graph.
4. Read path is available by default after schema/domain rollout:
   - snapshot/details HTTP + gRPC endpoints открываются вместе с доставкой схемы и доменного сервиса;
   - realtime stream открывается вместе с read path, без отдельных operator env-gates.
5. Enable core inline commands:
   - включить `discussion.create`, `work_item.create`, `discussion.formalize`, `stage.next_step.execute`, `command.retry_sync`.
   - `stage.next_step.execute` сначала включается только с path `pending_approval -> queued`, без bypass approval state.
6. Enable optional voice path:
   - только после отдельной owner decision и `CODEXK8S_MISSION_CONTROL_VOICE_ENABLED=true`.

## Как выполняются миграции при деплое
- Production deploy order remains mandatory:
  1. stateful dependencies ready
  2. migration job
  3. `control-plane`
  4. `worker`
  5. `api-gateway`
  6. `web-console`
- Concurrency control:
  - single migration runner + advisory lock (`goose` baseline)
- Failure policy:
  - если DDL или warmup fails, rollout stops before edge/frontend write-path exposure.

## Политика warmup/backfill
- Warmup source:
  - current GitHub issue/PR state
  - existing platform run/flow_events state
  - provider callbacks already persisted in platform domain
- Execution rules:
  - idempotent batches
  - restart-safe checkpoints
  - duplicate provider delivery ignored via unique keys
- Progress monitoring:
  - logs with processed entity counters
  - metrics for stale rows, rebuild duration and dedupe count

## Политика rollback
- Safe rollback before write-path enable:
  - drop/disable new routes and worker jobs
  - keep additive tables or drop them if no production data required
- Limited rollback after write-path enable:
  - new inline writes must be disabled first
  - provider-created issues/discussions/tasks are not reverted automatically
  - pending approval records and denied/expired approval audit are retained even if command execution stays disabled
  - command/timeline ledger remains preserved for audit and replay diagnosis
- Voice rollback:
  - disable `CODEXK8S_MISSION_CONTROL_VOICE_ENABLED`
  - keep `voice_candidates` data for audit, no destructive delete

## Что нельзя безопасно откатить
- Provider side effects already applied by accepted inline commands.
- Reconciled command history and audit evidence.
- User-visible relations created by formalization/promotion flows without compensating business command.

## Проверки
### Pre-migration checks
- absence of conflicting entity identity keys
- availability of feature flags:
  - `CODEXK8S_MISSION_CONTROL_ENABLED`
  - `CODEXK8S_MISSION_CONTROL_VOICE_ENABLED`
- provider credentials and webhook ingestion baseline healthy

### Post-migration verification
- warmup populated non-empty active-set projection for pilot projects
- snapshot latency within target
- duplicate webhook delivery does not create extra timeline entries
- command status transitions reach `reconciled` or `failed` deterministically
- approval-gated commands reach `pending_approval` before any side effect and only then progress to `queued`
- degraded fallback path works with realtime disabled

## Runtime impact / Migration impact
- Runtime impact (`run:design`): none.
- Migration impact (`run:dev`): moderate, additive tables/indexes + warmup job + feature-flagged route enablement.

## Operational notes
- If warmup lags, rollout may expose read-only snapshot path with `freshness_status=degraded`, but MUST NOT expose core inline commands yet.
- If realtime path is unstable, keep HTTP snapshot/details read path and disable WS stream independently.
- If voice path is unstable, disable only voice feature flag; core Mission Control rollout continues.
- If approval integration is unavailable, `stage.next_step.execute` MUST remain disabled rather than silently downgrading to direct label mutation.
