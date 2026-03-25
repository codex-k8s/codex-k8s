---
doc_id: MIG-S13-QG-0001
type: migrations-policy
title: "Quality Governance System — Migrations policy Sprint S13 Day 5"
status: in-review
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [466, 469, 470, 471, 476, 484, 488, 494, 512]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-494-migrations-policy"
---

# DB Migrations Policy: Quality Governance System

## TL;DR
- Подход: additive `expand -> domain-enable -> backfill/sweeps -> edge/ui-enable`.
- Владелец схемы/миграций: `services/internal/control-plane`.
- Миграции живут в `services/internal/control-plane/cmd/cli/migrations/*.sql`.
- Rollback ограничен после записи decision ledger и governance feedback: destructive delete не допускается.

## Размещение миграций и владелец схемы
- Schema owner: `services/internal/control-plane`.
- Все DDL и расширения для `Quality Governance System` выполняются только внутри owner service migrations directory:
  - `services/internal/control-plane/cmd/cli/migrations/*.sql`
- `worker`, `api-gateway`, `web-console` и `agent-runner` собственных DB migrations для Day5 scope не получают.

## Принципы
- Additive first:
  - новые tables/nullable columns появляются до включения любого live write path.
- Domain before background:
  - `control-plane` должен понимать package aggregate и projection refresh раньше, чем `worker` начнёт sweeps/backfill.
- Hidden-draft safety:
  - schema не должна вынуждать хранить raw draft content; только metadata and references.
- Bounded historical backfill:
  - backfill импортирует только evidence-backed open/active lineages и bounded recent release/postdeploy context;
  - hidden drafts, waivers и release decisions не фабрикуются retroactively.
- Visibility last:
  - `api-gateway` и `web-console` включаются после готовности package/projection model.

## Процесс миграции (план для `run:dev`)
1. Expand schema:
   - create `change_governance_packages`;
   - create `change_governance_internal_drafts`;
   - create `change_governance_waves`;
   - create `change_governance_evidence_blocks`;
   - create `change_governance_decision_records`;
   - create `change_governance_feedback_records`;
   - create `change_governance_projection_snapshots`;
   - create `change_governance_artifact_links`;
   - optionally add nullable `change_governance_package_id` to `flow_events`.
2. Index hardening:
   - queue/read indexes for packages;
   - latest draft partial unique index;
   - current projection partial unique index;
   - feedback severity/state indexes;
   - wave order uniqueness.
3. Enable `control-plane` domain path:
   - aggregation, hidden-draft recording, wave publication, decision ledger and projection refresh stay disabled while system setting `quality_governance_enabled=false`;
   - then enable live write path.
4. Enable `worker` sweeps and bounded backfill:
   - stale-gap detection;
   - late feedback ingestion;
   - bounded historical package seeding for active/open lineages only.
5. Enable edge/UI surfaces:
   - `api-gateway` DTO/casters and staff/private routes;
   - `web-console` queue/detail/gap views under typed platform setting `quality_governance_ui_enabled=true`.
6. Enable GitHub mirror if required:
   - comment mirror only after package projections are stable;
   - mirror remains optional and non-blocking.

## Как выполняются миграции при деплое
- Обязательный production order:
  1. stateful dependencies ready
  2. migration job
  3. `control-plane`
  4. `worker`
  5. `api-gateway`
  6. `web-console`
- Concurrency control:
  - single migration runner + advisory lock (`goose` baseline)
- Failure policy:
  - migration failure blocks rollout before service restart;
  - if `control-plane` cannot read new schema, `worker` and edge/frontend rollouts are blocked;
  - comment-mirror rollout never blocks canonical package state.

## Политика backfill
- Scope:
  - import open issue/PR lineages and bounded recent release/postdeploy contexts still relevant for governance review;
  - skip closed historical changes with no active evidence window.
- Rules:
  - backfill creates package root, artifact links and evidence blocks only from existing typed sources (`issues`, `PR`, `flow_events`, docs links, release/postdeploy evidence);
  - hidden draft records are never fabricated by backfill;
  - if historical wave map is absent, package remains visible with gap `semantic_mix` or `missing_evidence`, not with invented waves;
  - `high/critical` history without explicit waiver stays visibly gapped.
- Monitoring:
  - track counts of `backfill_created_packages_total`, `backfill_gap_seeded_total`, `backfill_failed_total`.

## Политика runtime settings
- `quality_governance_enabled`
  - typed platform setting в staff UI `/configuration/system-settings`;
  - controls live domain writes, hidden-draft ingestion, wave publication and projection refresh.
- future feedback/backfill toggle
  - должен добавляться в тот же typed platform settings catalog, когда stream `#522` подключит worker feedback/backfill.
- `quality_governance_ui_enabled`
  - controls staff/private route exposure and frontend rendering.
- `quality_governance_comment_mirror_enabled`
  - controls GitHub service-comment mirror only.
- Ordering:
  - UI/comment flags must never be enabled before domain flag;
  - feedback flag must not be enabled before domain flag.

## Политика rollback
- Safe rollback before feedback/UI enable:
  - disable `quality_governance_enabled` через staff UI `/configuration/system-settings`;
  - keep new schema/tables in place;
  - stop new package creation while leaving stored state readable.
- Limited rollback after decision recording:
  - disable UI and comment mirror first;
  - stop worker sweeps/backfill next;
  - keep decision/feedback history immutable for audit and manual handling.
- What cannot be rolled back safely:
  - decision ledger rows, residual-risk records and release decisions;
  - feedback records referencing live incidents/release outcomes;
  - projection versions referenced by audit and service comments.

## Проверки
### Pre-migration checks
- Existing `flow_events` and artifact link consumers remain compatible with nullable governance linkage.
- No service outside `control-plane` expects to write governance tables directly.
- Feature flags are declared with `CODEXK8S_` prefix and defaulted off.
- Deploy plan preserves mandatory order `migrations -> control-plane -> worker -> api-gateway -> web-console`.

### Post-migration verification
- Package queue indexes and current projection uniqueness are healthy.
- Hidden draft inserts store metadata only and reject attempts to persist raw draft payload.
- Wave map publication transitions package from `hidden_draft` to `wave_map_defined|waves_published` deterministically.
- `high/critical` package without waiver cannot pass release-readiness transition.
- Worker backfill creates only evidence-backed packages and surfaces gaps instead of inventing history.
- UI routes stay dark until `quality_governance_ui_enabled=true`.

## Operational notes
- If feedback/backfill jobs are noisy, disable `quality_governance_feedback_enabled` and keep package detail read-only.
- If comment mirror lags or fails, canonical owner/reviewer/operator surfaces stay in staff/private API; rollout is not blocked.
- If UI rollout lags, domain state may remain enabled with UI disabled.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): none.
- Migration impact (`run:dev`): moderate, additive tables + optional nullable audit linkage + staged enablement across domain/background/edge/UI.
