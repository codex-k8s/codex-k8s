---
doc_id: MIG-S12-GITHUB-RL-0001
type: migrations-policy
title: "GitHub API rate-limit resilience — Migrations policy Sprint S12 Day 5"
status: approved
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-15
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431, 500]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-420-migrations-policy"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# DB Migrations Policy: GitHub API rate-limit resilience

## TL;DR
- Подход: additive `expand -> domain-enable -> worker-enable -> runner-enable -> edge/ui-enable`.
- Владелец схемы/миграций: `services/internal/control-plane`.
- Миграции лежат в `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Rollback ограничен после включения runner handoff и worker sweeps: persisted waits/evidence сохраняются, destructive delete не допускается.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- Все DDL создаются внутри owner service migrations directory:
  - `services/internal/control-plane/cmd/cli/migrations/*.sql`
- `worker`, `agent-runner`, `api-gateway` и `web-console` не получают собственных DB migrations для S12.

## Принципы
- Additive first:
  - новые wait/evidence tables и enum/check expansions появляются до включения write path.
- Domain before orchestration:
  - `control-plane` must understand new wait states before `worker` or `agent-runner` send live traffic.
- Runner after worker:
  - agent handoff is enabled only after resume sweeps and dominant wait projection are already live.
- Visibility last:
  - edge/frontend rollout is intentionally last because staff surfaces can safely lag behind persisted domain state during short coordinated rollout.
- No `/rate_limit` polling:
  - migration and runtime logic do not introduce provider polling table or background sync against `GET /rate_limit`.

## Процесс миграции (план для `run:dev`)
1. Expand schema:
   - create `github_rate_limit_waits`;
   - create `github_rate_limit_wait_evidence`;
   - expand `agent_runs.status` with `waiting_backpressure`;
   - expand `agent_runs.wait_reason` with `github_rate_limit`;
   - expand `agent_runs.wait_target_kind` with `github_rate_limit_wait`;
   - expand `agent_sessions.wait_state` with `backpressure`.
2. Index hardening:
   - partial unique open wait per `(run_id, contour_kind)`;
   - partial unique dominant wait per `run_id`;
   - resume sweep index on `(state, resume_not_before)`;
   - evidence ordering index by `wait_id`.
3. Enable `control-plane` domain path:
   - classification, persistence, dominant wait election and read projection under `system_settings.github_rate_limit_wait_enabled=false` for read compatibility tests first;
   - then enable creation of live waits.
4. Enable `worker` sweeps:
   - resume scheduling and manual escalation for newly created waits;
   - comment mirror retry path included in the same wave.
5. Enable `agent-runner` handoff:
   - `ReportGitHubRateLimitSignal` RPC;
   - deterministic runner action `persist_session_and_exit_wait`;
   - resume payload handoff for `agent_session_resume`.
6. Enable edge/UI visibility:
   - `api-gateway` DTO/casters;
   - `web-console` dominant/related wait surfaces through typed contracts after backend rollout completes.

## Как выполняются миграции при деплое
- Production deploy order remains mandatory:
  1. stateful dependencies ready
  2. migration job
  3. `control-plane`
  4. `worker`
  5. `agent-runner`
  6. `api-gateway`
  7. `web-console`
- Concurrency control:
  - single migration runner + advisory lock (`goose` baseline)
- Failure policy:
  - if schema or index migration fails, rollout stops before any new service pod starts;
  - if `control-plane` cannot read the new schema, `worker` and `agent-runner` rollouts are blocked.

## Политика backfill
- Legacy rate-limit rows do not exist, so no historical data backfill is required for new wait tables.
- Existing wait rows for approval/interaction remain untouched.
- Compatibility normalization is limited to schema checks:
  - no existing `agent_runs` row may use `wait_reason=github_rate_limit` before domain cutover;
  - no existing `agent_sessions.wait_state` row may use `backpressure` before runner cutover.
- If any pilot wait is created during phased rollout, it must remain readable even when UI flag is off.

## Политика feature flags
- `system_settings.github_rate_limit_wait_enabled`
  - controls live creation and sweep execution of rate-limit waits.
- Staff/private visibility is no longer controlled отдельным env-only gate:
  - transport/UI rollout следует deploy order и typed DTO compatibility.

## Политика rollback
- Safe rollback before `agent-runner` handoff:
  - disable `system_settings.github_rate_limit_wait_enabled`;
  - keep tables and indexes in place;
  - do not create new waits;
  - existing hard-failure behavior remains effective.
- Limited rollback after runner handoff:
  - disable new wait creation and worker sweeps first;
  - stop UI exposure independently if needed;
  - keep persisted waits/evidence and session snapshots for audit and manual recovery.
- What cannot be rolled back safely:
  - audit evidence already written into `github_rate_limit_wait_evidence` and `flow_events`;
  - session snapshots prepared for deterministic agent resume;
  - semantic enum expansions already used by persisted rows.

## Continuity after `run:plan`
- Plan package Issue `#423` подтвердил этот rollout order как обязательный sequencing для waves `#425..#431`.
- Ни одна implementation wave не может обойти порядок `migrations -> control-plane -> worker -> agent-runner -> api-gateway -> web-console` или снять gate `#431` перед `run:qa`.

## Проверки
### Pre-migration checks
- `agent_runs` and `agent_sessions` contain no unknown wait enums before cutover.
- Existing `wait_target_kind`/`wait_target_ref` columns from S10 baseline are present and healthy.
- Feature flags are defined with `CODEXK8S_` prefix and defaulted off.
- Deployment plan confirms order `control-plane -> worker -> agent-runner -> edge/ui`.

### Post-migration verification
- Open wait partial unique indexes enforce one wait per `(run_id, contour_kind)`.
- Dominant wait partial unique index enforces one dominant wait per run.
- `control-plane` can create and read a pilot wait aggregate without `worker` enabled.
- `worker` sweep can transition `open -> auto_resume_scheduled -> resolved/manual_action_required` idempotently.
- `agent-runner` signal callback persists wait + session snapshot and exits without local retry loop.
- Staff routes and realtime stream render typed wait projection only after UI flag enable.

## Operational notes
- If service-comment mirror fails because platform contour is saturated, rollout does not fail:
  - canonical visibility remains staff/private surfaces;
  - worker retries comment sync later.
- If `worker` sweeps are unstable, disable the domain flag and keep waits visible/read-only for operator diagnosis.
- If UI rollout lags, keep domain flag on and UI flag off; internal state remains valid.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): none.
- Migration impact (`run:dev`): moderate, additive tables/indexes + enum/check expansion + staged service rollout.
