---
doc_id: DM-S16-MISSION-CONTROL-0001
type: data-model
title: "Mission Control graph workspace — Data model Sprint S16 Day 5"
status: superseded
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-25
related_issues: [480, 490, 492, 496, 510, 516, 519, 561, 562, 563]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-519-data-model"
---

# Data Model: Mission Control graph workspace

## TL;DR
- 2026-03-25 issue `#561` перевела этот data model в historical superseded state.
- Node kinds `discussion/work_item/run/pull_request`, continuity-gap schema Sprint S16 и reuse path из этого файла больше не являются текущим Mission Control baseline.
- Новый agreed Wave 1 baseline после rethink: `Issue`, `PR`, `Run`; соответствующий data model должен быть заново определён после `#562`.

## Entities
### Entity: `mission_control_entities`
- Purpose: canonical graph-node storage for workspace projection under `control-plane` ownership.
- Important invariants:
  - `project_id + entity_kind + entity_external_key` remains unique.
  - public transport identifier stays stable and maps 1:1 to `entity_external_key`.
  - target Wave 1 transport emits only `discussion|work_item|run|pull_request`; any `agent` rows are migration-only and must not leak through S16 transport.
  - continuity completeness is never guessed in UI; it is derived from persisted gap rows and summarized in `continuity_status`.
  - bounded recent closed context is explicit through `coverage_class`, not inferred by stale UI logic.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | internal row id |
| project_id | uuid | no |  | fk -> projects | tenant boundary |
| entity_kind | text | no |  | check(discussion/work_item/run/pull_request/agent) | `agent` kept only for migration/cleanup window |
| entity_external_key | text | no |  | unique(project_id, entity_kind, entity_external_key) | public node id |
| provider_kind | text | no | `github` | check(github/platform) | provider vs platform-owned node |
| provider_url | text | yes |  |  | provider deep-link target |
| title | text | no |  |  | graph/list label |
| active_state | text | no |  | check(working/waiting/blocked/review/recent_critical_updates/archived) | active-state preset source |
| sync_status | text | no | `synced` | check(synced/pending_sync/failed/degraded) | provider/platform freshness summary |
| continuity_status | text | no | `complete` | check(complete/missing_run/missing_pull_request/missing_follow_up_issue/stale_provider/out_of_scope) | derived summary from gaps + watermarks |
| coverage_class | text | no | `open_primary` | check(open_primary/recent_closed_context/out_of_scope) | makes bounded recent history explicit |
| projection_version | bigint | no | 1 | check(>=1) | optimistic concurrency for preview/commands |
| card_payload | jsonb | no | '{}'::jsonb |  | graph/list node summary slice; reuse S9 column to keep migration additive |
| detail_payload | jsonb | no | '{}'::jsonb |  | drawer payload by node kind |
| last_timeline_at | timestamptz | yes |  |  | used for ordering and root freshness |
| provider_updated_at | timestamptz | yes |  |  | last seen provider change |
| projected_at | timestamptz | no | now() |  | last node projection write |
| stale_after | timestamptz | yes |  |  | node-level freshness threshold |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: `mission_control_relations`
- Purpose: persisted graph edges between workspace nodes.
- Important invariants:
  - duplicate edge by `(source_entity_id, relation_kind, target_entity_id)` is forbidden.
  - edge direction is canonical and not inverted by transport layer.
  - edges required only for graph integrity may exist even when one endpoint is `secondary_dimmed`, but they still respect the same canonical relation kinds.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| project_id | uuid | no |  | fk -> projects | |
| source_entity_id | bigint | no |  | fk -> mission_control_entities | |
| relation_kind | text | no |  | check(formalized_from/spawned_run/produced_pull_request/continues_with/blocked_by/related_to/tracked_by_command) | closed Wave 1 edge set |
| target_entity_id | bigint | no |  | fk -> mission_control_entities | |
| source_kind | text | no | `platform` | check(platform/provider/command) | provenance of the edge |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: `mission_control_timeline_entries`
- Purpose: drawer activity stream for provider comments/reviews, platform flow events and command lifecycle.
- Important invariants:
  - `entry_external_key` remains unique for `(project_id, source_kind)`.
  - comments/chat stay drawer entities and never become graph nodes.
  - ordering is canonical at storage layer, not reconstructed in UI.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| project_id | uuid | no |  | fk -> projects | |
| entity_id | bigint | no |  | fk -> mission_control_entities | owner node |
| source_kind | text | no |  | check(provider/platform/command) | |
| entry_external_key | text | no |  | unique(project_id, source_kind, entry_external_key) | dedupe key |
| command_id | uuid | yes |  | fk -> mission_control_commands | nullable for non-command events |
| summary | text | no |  |  | short drawer line |
| body_markdown | text | yes |  |  | provider or platform body |
| payload | jsonb | no | '{}'::jsonb |  | typed extras by source kind |
| occurred_at | timestamptz | no |  | index | canonical ordering |
| provider_url | text | yes |  |  | deep-link for provider-originated entries |
| is_read_only | bool | no | true |  | provider entries remain read-only |
| created_at | timestamptz | no | now() |  | |

### Entity: `mission_control_commands`
- Purpose: single command ledger for platform-safe actions, including graph-driven `stage.next_step.execute`.
- Important invariants:
  - no second command path is introduced for graph workspace.
  - `business_intent_key` remains semantic dedupe anchor.
  - preview result is not a command row; side effects start only after explicit submit.
  - lease columns remain necessary for worker claim/reconcile flow.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | public command id |
| project_id | uuid | no |  | fk -> projects | |
| command_kind | text | no |  | check(discussion.create/work_item.create/discussion.formalize/stage.next_step.execute/command.retry_sync) | retained from S9 |
| target_entity_id | bigint | yes |  | fk -> mission_control_entities | target node for existing actions |
| actor_id | text | no |  |  | principal snapshot |
| business_intent_key | text | no |  | unique(project_id, business_intent_key) | dedupe anchor |
| correlation_id | text | no |  | unique | end-to-end audit correlation |
| status | text | no | `accepted` | check(accepted/pending_approval/queued/pending_sync/reconciled/failed/blocked/cancelled) | |
| failure_reason | text | yes |  | check(provider_error/policy_denied/projection_stale/duplicate_intent/timeout/approval_denied/approval_expired/unknown) | |
| approval_request_id | uuid | yes |  |  | nullable |
| approval_state | text | no | `not_required` | check(not_required/pending/approved/denied/expired) | |
| approval_requested_at | timestamptz | yes |  |  | |
| approval_decided_at | timestamptz | yes |  |  | |
| payload | jsonb | no | '{}'::jsonb |  | typed command payload including expected projection version |
| result_payload | jsonb | no | '{}'::jsonb |  | typed ack/result |
| provider_delivery_ids | jsonb | no | '[]'::jsonb |  | observed provider events |
| lease_owner | text | yes |  |  | worker lease owner |
| lease_until | timestamptz | yes |  |  | worker lease deadline |
| requested_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |
| reconciled_at | timestamptz | yes |  |  | |

### Entity: `mission_control_continuity_gaps`
- Purpose: persisted completeness gaps for graph lineage and next-step eligibility.
- Important invariants:
  - missing PR or missing follow-up issue is stored as a row, not a computed-only banner.
  - one open gap of the same `gap_kind` per `(project_id, subject_entity_id)` is unique.
  - gap resolution references actual resulting node when available.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | public transport may expose opaque gap id |
| project_id | uuid | no |  | fk -> projects | |
| subject_entity_id | bigint | no |  | fk -> mission_control_entities | node where continuity breaks |
| gap_kind | text | no |  | check(missing_run/missing_pull_request/missing_follow_up_issue/provider_out_of_scope/provider_stale/orphan_node) | |
| severity | text | no | `warning` | check(blocking/warning/info) | |
| status | text | no | `open` | check(open/resolved/deferred) | |
| expected_entity_kind | text | yes |  | check(discussion/work_item/run/pull_request) | nullable for provider-scope gaps |
| expected_stage_label | text | yes |  |  | expected downstream stage when relevant |
| resolution_entity_id | bigint | yes |  | fk -> mission_control_entities | nullable until resolved |
| resolution_hint | text | yes |  |  | short operator-facing guidance |
| payload | jsonb | no | '{}'::jsonb |  | typed extras, e.g. missing thread refs |
| detected_at | timestamptz | no | now() |  | |
| resolved_at | timestamptz | yes |  |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: `mission_control_workspace_watermarks`
- Purpose: workspace-level typed freshness/coverage markers for provider mirror, projection freshness and launch policy generation.
- Important invariants:
  - scope remains workspace-level in Wave 1; node-level freshness stays in `mission_control_entities`.
  - bounded recent closed history must surface explicit window bounds.
  - newest row per `(project_id, watermark_kind)` is the effective workspace watermark.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| project_id | uuid | no |  | fk -> projects | |
| watermark_kind | text | no |  | check(provider_freshness/provider_coverage/graph_projection/launch_policy) | |
| status | text | no |  | check(fresh/stale/degraded/out_of_scope) | |
| summary | text | no |  |  | user-visible explanation |
| window_started_at | timestamptz | yes |  |  | bounded recent closed start |
| window_ended_at | timestamptz | yes |  |  | bounded recent closed end / observed upper bound |
| observed_at | timestamptz | no | now() |  | when watermark was computed |
| payload | jsonb | no | '{}'::jsonb |  | typed extras like provider cursor/checkpoint |
| created_at | timestamptz | no | now() |  | append-only history |

## JSONB payload variant mapping
### `mission_control_entities.detail_payload`
| `entity_kind` | Stored variant |
|---|---|
| `discussion` | `MissionControlDiscussionNodeDetails` |
| `work_item` | `MissionControlWorkItemNodeDetails` |
| `run` | `MissionControlRunNodeDetails` |
| `pull_request` | `MissionControlPullRequestNodeDetails` |
| `agent` | Transitional legacy payload, never exposed by S16 transport |

### `mission_control_entities.card_payload`
| Purpose | Stored variant |
|---|---|
| graph/list summary | `MissionControlNodeSummaryPayload` |

### `mission_control_timeline_entries.payload`
| `source_kind` | Stored variant |
|---|---|
| `provider` | `MissionControlProviderActivityPayload` |
| `platform` | `MissionControlPlatformActivityPayload` |
| `command` | `MissionControlCommandActivityPayload` |

### `mission_control_commands.payload`
| `command_kind` | Stored variant |
|---|---|
| `discussion.create` | `MissionControlDiscussionCreatePayload` |
| `work_item.create` | `MissionControlWorkItemCreatePayload` |
| `discussion.formalize` | `MissionControlDiscussionFormalizePayload` |
| `stage.next_step.execute` | `MissionControlStageNextStepPayload` |
| `command.retry_sync` | `MissionControlRetrySyncPayload` |

### `mission_control_continuity_gaps.payload`
| `gap_kind` | Stored variant |
|---|---|
| `missing_run` | `MissionControlMissingRunGapPayload` |
| `missing_pull_request` | `MissionControlMissingPullRequestGapPayload` |
| `missing_follow_up_issue` | `MissionControlMissingFollowUpGapPayload` |
| `provider_out_of_scope` | `MissionControlCoverageGapPayload` |
| `provider_stale` | `MissionControlStalenessGapPayload` |
| `orphan_node` | `MissionControlOrphanNodeGapPayload` |

## Relationships
- `projects` 1:N `mission_control_entities`
- `mission_control_entities` N:N `mission_control_entities` via `mission_control_relations`
- `mission_control_entities` 1:N `mission_control_timeline_entries`
- `mission_control_entities` 1:N `mission_control_continuity_gaps`
- `mission_control_entities` 1:N `mission_control_commands`
- `mission_control_commands` 1:N `mission_control_timeline_entries`
- `projects` 1:N `mission_control_workspace_watermarks`
- External source dependencies:
  - provider mirror state from foundation issue `#480`;
  - `agent_runs`, `flow_events`, `links` supply run lineage and backfill evidence.

## Critical indexes and queries
- Query: workspace root groups by preset + search
  - Indexes:
    - `(project_id, active_state, continuity_status, projected_at desc)`
    - `(project_id, coverage_class, projected_at desc)`
    - GIN on `card_payload`
- Query: node details by public id
  - Index:
    - unique `(project_id, entity_kind, entity_external_key)`
- Query: edge lookup around a node
  - Indexes:
    - `(source_entity_id, relation_kind, updated_at desc)`
    - `(target_entity_id, relation_kind, updated_at desc)`
- Query: open blocking gaps
  - Indexes:
    - `(project_id, status, severity, detected_at desc)`
    - unique partial on `(project_id, subject_entity_id, gap_kind)` where `status='open'`
- Query: latest workspace watermarks
  - Index:
    - `(project_id, watermark_kind, observed_at desc)`
- Query: command lookup and worker claim
  - Indexes:
    - unique `(project_id, business_intent_key)`
    - `(status, lease_until, updated_at desc)` partial for claimable commands

## Retention and lifecycle policy
- `mission_control_entities`:
  - active/open and recent-closed-context rows stay until provider mirror or graph truth evicts them from bounded scope;
  - no invisible silent delete for recent closed context without updated coverage watermark.
- `mission_control_relations`:
  - append-or-replace by owner use-case; obsolete edges are removed only when graph truth changes deterministically.
- `mission_control_timeline_entries`:
  - append-only except duplicate-safe no-op.
- `mission_control_continuity_gaps`:
  - resolved/deferred rows stay for audit and rollout diagnosis.
- `mission_control_workspace_watermarks`:
  - append-only history; latest row per kind is effective.

## Domain invariants
- `run` is a platform-native node and must exist before a PR edge can be considered complete in Wave 1 lineage.
- Missing linked follow-up issue after a stage is always a persisted `missing_follow_up_issue` gap.
- `provider_out_of_scope` and `provider_stale` are not equivalent:
  - `out_of_scope` means entity is outside bounded policy;
  - `stale` means entity is inside policy window but freshness has degraded.
- `secondary_dimmed` is a projection concern and never changes canonical node ownership or edge semantics.
- `stage.next_step.execute` cannot move from preview to queued/reconciled without rechecking `projection_version` and approval state.

## Ownership and write-path
- `control-plane` owns:
  - schema;
  - repositories;
  - graph classification;
  - continuity-gap lifecycle;
  - workspace watermarks;
  - command admission and state transitions.
- `worker` owns:
  - provider sync and recent-closed-history backfill;
  - projection rebuild execution through `control-plane` use-cases.
- `api-gateway` and `web-console` never persist canonical Mission Control state.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): none; documents only.
- Migration impact (`run:dev`):
  - additive columns on `mission_control_entities` and `mission_control_relations`;
  - new tables `mission_control_continuity_gaps`, `mission_control_workspace_watermarks`;
  - backfill from existing Mission Control projection + runtime/provider evidence before transport cutover.

## Open questions
- Should `mission_control_workspace_watermarks` keep full append-only history in Wave 1, or is “latest per kind + audit in `flow_events`” sufficient for the first rollout?

## Апрув
- request_id: `owner-2026-03-16-issue-519-data-model`
- Решение: pending
- Комментарий: требуется owner review schema evolution path и handover в `run:plan`.
