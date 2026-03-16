---
doc_id: MIG-S16-MISSION-CONTROL-0001
type: migrations-policy
title: "Mission Control graph workspace — DB migrations policy Sprint S16 Day 5"
status: in-review
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510, 516, 519]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-519-migrations-policy"
---

# DB Migrations Policy: Mission Control graph workspace

## TL;DR
- Подход: additive expand -> shadow backfill -> read switch -> command preview exposure -> cleanup.
- Владелец схемы/миграций: `services/internal/control-plane`.
- Новые миграции лежат в `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Rollback ограничен после cutover на graph-first transport, но destructive schema rollback не требуется: additive schema can stay while app images roll back.
- Existing Sprint S9 Mission Control foundation is not discarded; it is evolved in place with new continuity-gap and watermark storage.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- All DDL stays in:
  - `services/internal/control-plane/cmd/cli/migrations/*.sql`
- Shared DB without owner is forbidden:
  - `worker`, `api-gateway` and `web-console` do not get their own mission-control migrations.

## Principles
- Reuse before duplicate:
  - existing Mission Control nodes/edges/activity/commands tables stay under one owner and are extended rather than replaced by a parallel write path.
- Shadow first:
  - graph backfill and continuity-gap generation run before transport cutover.
- Read before write:
  - graph snapshot/details paths are validated before `launch-preview` becomes a default CTA.
- No env-only feature gates:
  - rollout relies on candidate validation, deploy order and owner-governed PR review, not on a new runtime env flag.
- Cleanup last:
  - old `agent` node materialization and obsolete S9-only route/docs cleanup happen after graph cutover is stable.

## Migration process (target for `run:dev`)
1. Expand schema:
   - alter `mission_control_entities` to support `run` nodes and add `continuity_status`, `coverage_class`;
   - alter `mission_control_relations` to allow graph lineage edge kinds `spawned_run`, `produced_pull_request`, `continues_with`, `related_to`;
   - create `mission_control_continuity_gaps`;
   - create `mission_control_workspace_watermarks`;
   - add indexes for gap lookup, watermark lookup and graph root queries.
2. Shadow backfill:
   - backfill `run` nodes from `agent_runs`, `flow_events`, `links` and provider mirror evidence;
   - compute continuity gaps and workspace watermarks from the same bounded coverage window as issue `#480`;
   - stop treating `agent` nodes as canvas materialization target, but keep legacy rows until cleanup.
3. Candidate parity checks:
   - compare graph root count and node coverage against provider mirror for `all open Issues/PR + bounded recent closed history`;
   - verify no `agent` node leaks into transport DTO;
   - verify every `missing_pull_request` or `missing_follow_up_issue` is represented by a persisted gap row.
4. Read-side cutover:
   - switch OpenAPI/gRPC/read-model code to `/workspace`, `/nodes`, `/activity`, `/realtime`;
   - deploy `control-plane` and `worker` before exposing new transport in `api-gateway`;
   - deploy `web-console` last.
5. Launch-preview exposure:
   - enable graph preview route only after read-side parity passes;
   - keep actual mutations on existing `POST /commands` path with `stage.next_step.execute`.
6. Cleanup:
   - remove old S9 dashboard-only docs/routes from active contract;
   - create a later cleanup migration if `agent` rows and legacy edge kinds are no longer needed after candidate soak.

## How migrations run during deploy
- Mandatory rollout order stays:
  1. stateful dependencies ready
  2. migration job
  3. `control-plane`
  4. `worker`
  5. `api-gateway`
  6. `web-console`
- Concurrency control:
  - one migration runner with advisory lock (`goose` baseline).
- Failure policy:
  - if DDL or shadow backfill fails, rollout stops before `api-gateway` and `web-console` switch to graph-first contract.

## Backfill policy
- Source inputs:
  - provider mirror foundation from issue `#480`;
  - current `mission_control_*` rows from Sprint S9;
  - `agent_runs`, `flow_events`, `links` for run lineage and continuity evidence.
- Execution rules:
  - idempotent batches by `(project_id, entity_kind, entity_external_key)` and unique open gap keys;
  - restart-safe checkpoints;
  - no page-load live fetch as primary recovery path.
- Progress monitoring:
  - processed nodes/edges/gaps counters in logs;
  - metrics for watermark age, unresolved blocking gaps and backfill duration.

## Rollback policy
- Safe rollback before read-side cutover:
  - keep additive schema;
  - redeploy previous app version;
  - disable shadow backfill jobs if they are the failure source.
- Safe rollback after read-side cutover but before merge:
  - redeploy previous `web-console` and `api-gateway`, then previous `worker` and `control-plane`;
  - retain additive schema and backfilled rows for audit and future replay.
- Limited rollback after merge:
  - app-image rollback remains the primary mechanism;
  - no destructive down migration for graph tables is required;
  - provider side effects already executed by existing `mission_control_commands` cannot be reverted automatically.

## What cannot be safely rolled back
- Provider mutations already reconciled through `stage.next_step.execute`.
- Historical continuity-gap evidence already written for audit.
- Watermark history rows that were already used in operator diagnosis.

## Checks
### Pre-migration checks
- provider mirror coverage for open Issues/PR and recent closed window is healthy;
- candidate namespace/build lineage is ready for lockstep backend/frontend validation;
- no unresolved drift between existing next-step policy and graph-preview design docs.

### Post-migration verification
- `run` nodes are present in projection and `agent` nodes are absent from transport;
- graph snapshot returns root groups, edges, gaps and workspace watermarks for pilot projects;
- bounded recent closed history appears only as `secondary_dimmed` context with matching watermark window;
- preview route returns typed continuity effect without writing commands;
- command lifecycle still flows through `accepted -> pending_approval|queued -> pending_sync -> reconciled|failed`.

## Operational notes
- If shadow backfill lags, do not expose graph-first transport yet; keep rollout blocked at backend wave.
- If provider mirror freshness degrades, show `provider_coverage`/`provider_freshness` watermarks rather than widening scope heuristically.
- If preview route drifts from existing next-step policy, disable preview exposure and keep command submission inaccessible from graph shell until reconciled.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): none.
- Migration impact (`run:dev`): moderate; additive schema evolution plus backfill and transport cutover, but no new service or new DB owner.

## Open questions
- Нужен ли отдельный owner-managed cleanup issue на SQL/transport cleanup после candidate soak, или можно включить cleanup в последнюю wave `run:dev` без риска для rollback?

## Апрув
- request_id: `owner-2026-03-16-issue-519-migrations-policy`
- Решение: pending
- Комментарий: требуется owner review rollout/backfill/rollback discipline и handover в `run:plan`.
