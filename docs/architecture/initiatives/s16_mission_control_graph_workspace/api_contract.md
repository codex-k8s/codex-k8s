---
doc_id: API-S16-MISSION-CONTROL-0001
type: api-contract
title: "Mission Control graph workspace — API contract Sprint S16 Day 5"
status: in-review
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510, 516, 519]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-519-api-contract"
---

# API Contract: Mission Control graph workspace

## TL;DR
- Тип API: staff/private REST + internal gRPC + WebSocket realtime stream.
- Аутентификация: staff JWT + project RBAC; preview stays read-only, mutating actions remain in `control-plane` command admission.
- Версионирование: `/api/v1` for HTTP and `v1` package for gRPC/realtime payloads.
- Основные операции: workspace snapshot, node details, node activity, launch preview, command submit/status, realtime graph delta.
- Compatibility decision: S16 supersedes Sprint S9 `dashboard/entities/timeline` contract in the same Mission Control namespace; no parallel long-lived HTTP surface is introduced.

## Спецификации (source of truth)
- OpenAPI (обновляется в `run:dev`): `services/external/api-gateway/api/server/api.yaml`
- gRPC proto (обновляется в `run:dev`): `proto/codexk8s/controlplane/v1/controlplane.proto`
- Transport mapping requirements:
  - HTTP DTO: `services/external/api-gateway/internal/transport/http/models`
  - HTTP casters: `services/external/api-gateway/internal/transport/http/casters`
  - gRPC DTO/casters: `services/internal/control-plane/internal/transport/grpc/{models,casters}`

## Staff HTTP endpoints (design baseline)
| Operation | Method | Path | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Get workspace snapshot | GET | `/api/v1/staff/mission-control/workspace` | staff JWT | n/a | Query: `view_mode`, `state_preset`, `search`, `cursor`, `root_limit` |
| Get node details | GET | `/api/v1/staff/mission-control/nodes/{node_kind}/{node_public_id}` | staff JWT | n/a | Returns drawer payload, gaps, watermarks and launch surfaces |
| List node activity | GET | `/api/v1/staff/mission-control/nodes/{node_kind}/{node_public_id}/activity` | staff JWT | n/a | Cursor pagination over provider/platform activity stream |
| Preview launch | POST | `/api/v1/staff/mission-control/launch-preview` | staff JWT | n/a | Read-only preview for `stage.next_step.execute` and continuity effect |
| Submit command | POST | `/api/v1/staff/mission-control/commands` | staff JWT | `Idempotency-Key` | Existing command ledger is retained for platform-safe actions |
| Get command status | GET | `/api/v1/staff/mission-control/commands/{command_id}` | staff JWT | n/a | Poll/read fallback when realtime unavailable |
| Connect realtime stream | GET | `/api/v1/staff/mission-control/realtime` | staff JWT | n/a | WebSocket upgrade; requires `resume_token` from latest snapshot |

## Query semantics
- `view_mode`:
  - `graph`
  - `list`
- `state_preset`:
  - `working`
  - `waiting`
  - `blocked`
  - `review`
  - `recent_critical_updates`
  - `all_active`
- Fixed Wave 1 filters are not parameterized:
  - `open_scope` is always `open_only`;
  - `assignment_scope` is always `assigned_to_me_or_unassigned` relative to current principal.
- `cursor` paginates root groups, not individual nodes.
- `search` applies after fixed filters over typed indexed fields: title, public ids, stage label, run id, branch head/base and selected labels.

## Internal gRPC methods (design baseline)
| RPC | Request | Response | Error mapping |
|---|---|---|---|
| `GetMissionControlWorkspace` | `GetMissionControlWorkspaceRequest` | `GetMissionControlWorkspaceResponse` | `invalid_argument`, `forbidden`, `internal` |
| `GetMissionControlNode` | `GetMissionControlNodeRequest` | `MissionControlNodeDetails` | `not_found`, `forbidden` |
| `ListMissionControlNodeActivity` | `ListMissionControlNodeActivityRequest` | `ListMissionControlNodeActivityResponse` | `not_found`, `forbidden` |
| `PreviewMissionControlLaunch` | `PreviewMissionControlLaunchRequest` | `MissionControlLaunchPreview` | `failed_precondition`, `forbidden`, `conflict` |
| `SubmitMissionControlCommand` | `SubmitMissionControlCommandRequest` | `MissionControlCommandState` | `invalid_argument`, `conflict`, `failed_precondition`, `forbidden` |
| `GetMissionControlCommand` | `GetMissionControlCommandRequest` | `MissionControlCommandState` | `not_found`, `forbidden` |
| `OpenMissionControlRealtime` | `OpenMissionControlRealtimeRequest` | stream `MissionControlRealtimeEnvelope` | `failed_precondition`, `unauthorized` |

## Key DTOs
### MissionControlWorkspaceSnapshot DTO
- `snapshot_id` (string, opaque)
- `view_mode` (`graph|list`)
- `generated_at` (RFC3339)
- `resume_token` (string)
- `effective_filters`:
  - `open_scope` (`open_only`)
  - `assignment_scope` (`assigned_to_me_or_unassigned`)
  - `state_preset`
  - `search` (string, optional)
- `summary`:
  - `root_count`
  - `node_count`
  - `blocking_gap_count`
  - `warning_gap_count`
  - `recent_closed_context_count`
  - `working_count`
  - `waiting_count`
  - `blocked_count`
  - `review_count`
  - `recent_critical_updates_count`
- `workspace_watermarks[]` (`MissionControlWorkspaceWatermark`)
- `root_groups[]` (`MissionControlRootGroup`)
- `nodes[]` (`MissionControlNode`)
- `edges[]` (`MissionControlEdge`)
- `next_root_cursor` (string, optional)

### MissionControlRootGroup DTO
- `root_node_kind`
- `root_node_public_id`
- `root_title`
- `node_refs[]`
- `has_blocking_gap`
- `latest_activity_at` (RFC3339)

### MissionControlNode DTO
- `node_kind` (`discussion|work_item|run|pull_request`)
- `node_public_id` (string)
- `title` (string)
- `visibility_tier` (`primary|secondary_dimmed`)
- `active_state` (`working|waiting|blocked|review|recent_critical_updates|archived`)
- `continuity_status` (`complete|missing_run|missing_pull_request|missing_follow_up_issue|stale_provider|out_of_scope`)
- `coverage_class` (`open_primary|recent_closed_context|out_of_scope`)
- `provider_reference` (`MissionControlProviderReference`, optional)
- `root_node_public_id` (string)
- `column_index` (int32)
- `last_activity_at` (RFC3339, optional)
- `has_blocking_gap` (bool)
- `badges[]` (`continuity_gap`, `provider_stale`, `recent_closed_context`, `waiting_mcp`, `review_required`)
- `projection_version` (int64)

### MissionControlEdge DTO
- `edge_kind` (`formalized_from|spawned_run|produced_pull_request|continues_with|blocked_by|related_to|tracked_by_command`)
- `source_node_kind`
- `source_node_public_id`
- `target_node_kind`
- `target_node_public_id`
- `visibility_tier` (`primary|secondary_dimmed`)
- `source_of_truth` (`platform|provider|command`)
- `is_primary_path` (bool)

### MissionControlWorkspaceWatermark DTO
- `watermark_kind` (`provider_freshness|provider_coverage|graph_projection|launch_policy`)
- `status` (`fresh|stale|degraded|out_of_scope`)
- `summary` (string)
- `observed_at` (RFC3339)
- `window_started_at` (RFC3339, optional)
- `window_ended_at` (RFC3339, optional)

### MissionControlNodeDetails DTO
- `node` (`MissionControlNode`)
- `detail_payload` (typed union by `node_kind`)
- `adjacent_nodes[]`
- `adjacent_edges[]`
- `continuity_gaps[]` (`MissionControlContinuityGap`)
- `node_watermarks[]` (`MissionControlWorkspaceWatermark`)
- `activity_preview[]` (`MissionControlActivityEntry`)
- `launch_surfaces[]` (`MissionControlLaunchSurface`)
- `provider_deep_links[]` (`MissionControlProviderDeepLink`)

### Node detail payload variants
| `node_kind` | DTO variant | Required fields |
|---|---|---|
| `discussion` | `MissionControlDiscussionNodeDetails` | `discussion_kind`, `status`, `author`, `participant_count`, `latest_comment_excerpt`, `formalization_target_refs[]` |
| `work_item` | `MissionControlWorkItemNodeDetails` | `repository_full_name`, `issue_number`, `stage_label`, `labels[]`, `assignees[]`, `linked_run_refs[]`, `linked_follow_up_refs[]`, `last_provider_sync_at` |
| `run` | `MissionControlRunNodeDetails` | `run_id`, `agent_key`, `run_status`, `runtime_mode`, `trigger_label`, `build_ref`, `candidate_namespace`, `started_at`, `finished_at`, `linked_pull_request_refs[]`, `produced_issue_refs[]` |
| `pull_request` | `MissionControlPullRequestNodeDetails` | `repository_full_name`, `pull_request_number`, `branch_head`, `branch_base`, `merge_state`, `review_decision`, `checks_summary`, `linked_issue_refs[]`, `linked_run_ref` |

### MissionControlActivityEntry DTO
- `entry_id`
- `node_kind`
- `node_public_id`
- `source_kind` (`provider|platform|command`)
- `source_ref`
- `occurred_at` (RFC3339)
- `summary`
- `body_markdown` (optional)
- `provider_url` (optional)
- `is_read_only` (bool)

### MissionControlContinuityGap DTO
- `gap_id`
- `gap_kind` (`missing_run|missing_pull_request|missing_follow_up_issue|provider_out_of_scope|provider_stale|orphan_node`)
- `severity` (`blocking|warning|info`)
- `status` (`open|resolved|deferred`)
- `subject_node_kind`
- `subject_node_public_id`
- `expected_node_kind` (optional)
- `expected_stage_label` (optional)
- `detected_at` (RFC3339)
- `resolved_at` (RFC3339, optional)
- `resolution_hint` (optional string)

### MissionControlLaunchSurface DTO
- `action_kind` (`preview_next_stage|open_linked_pull_request|open_linked_follow_up_issue|open_provider_context|inspect_run_context`)
- `presentation` (`primary|secondary|link`)
- `approval_requirement` (`none|owner_review`)
- `blocked_reason` (optional)
- `command_template` (optional `MissionControlStageNextStepTemplate`)

### MissionControlStageNextStepTemplate DTO
- `thread_kind` (`issue|pull_request`)
- `thread_number` (int32)
- `target_label`
- `removed_labels[]`
- `display_variant`
- `approval_requirement`
- `expected_gap_ids[]`

### Preview transport
#### Preview request
- `node_kind`
- `node_public_id`
- `thread_kind`
- `thread_number`
- `target_label`
- `removed_labels[]`
- `expected_projection_version`

#### Preview response
- `preview_id`
- `approval_requirement`
- `label_diff`:
  - `removed_labels[]`
  - `added_labels[]`
  - `final_labels[]`
- `continuity_effect`:
  - `resolved_gap_ids[]`
  - `remaining_gap_ids[]`
  - `resulting_node_refs[]`
  - `provider_redirects[]`
- `blocking_reason` (optional)

### Command request/response
- Existing write-path is retained with graph-aware refs:
  - `command_kind` (`discussion.create|work_item.create|discussion.formalize|stage.next_step.execute|command.retry_sync`)
  - `target_node_kind`
  - `target_node_public_id`
  - `business_intent_key`
  - `expected_projection_version`
  - typed payload per `command_kind`
- `MissionControlCommandState` remains the source of truth for command lifecycle:
  - `accepted`
  - `pending_approval`
  - `queued`
  - `pending_sync`
  - `reconciled`
  - `failed`
  - `blocked`
  - `cancelled`

### Realtime envelope
- `event_kind`:
  - `connected`
  - `delta`
  - `invalidate`
  - `stale`
  - `degraded`
  - `resync_required`
  - `heartbeat`
  - `error`
- `delta` payload includes:
  - `delta_nodes[]`
  - `delta_edges[]`
  - `delta_gaps[]`
  - `delta_workspace_watermarks[]`
  - `changed_command_ids[]`

## Validation and guardrails
- `node_kind` is closed enum `discussion|work_item|run|pull_request`; `agent` is forbidden in transport for Wave 1.
- `secondary_dimmed` is valid only when:
  - node is required for graph integrity between primary nodes; or
  - node represents recent closed context inside the bounded coverage window.
- Preview route is read-only and must not create commands, labels or provider side effects.
- `stage.next_step.execute` may be submitted only after caller received preview or consciously bypassed it via explicit UI policy; backend still revalidates all policy and continuity invariants.
- `open_provider_context` and other provider deep links are transport objects only, never mutation fallbacks.
- `recent_closed_context` nodes must always ship with watermark status that explains why they are visible.

## Error model
- Canonical domain codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- Typical mappings:
  - preview against stale projection -> `failed_precondition`;
  - node outside bounded coverage window -> `failed_precondition`;
  - unknown `node_kind` or `state_preset` -> `invalid_argument`;
  - missing RBAC or provider-unsafe action -> `forbidden`.
- Rate limits:
  - standard staff API rate limit for snapshot/details;
  - stricter limit for launch preview and command submit.

## Backward compatibility
- S16 does not preserve Sprint S9 `dashboard/entities/timeline` DTO contract.
- Rationale:
  - repo is early-stage and same frontend/backend ship in one monorepo;
  - graph workspace needs explicit `run` node, gaps, watermarks and graph-first filters that do not fit old board contract cleanly.
- Compatibility strategy:
  - rollout stays lockstep inside one PR and one candidate lineage;
  - no parallel long-lived transport namespace is introduced.

## Наблюдаемость
- API logs:
  - `mission_control.api.workspace`
  - `mission_control.api.node_details`
  - `mission_control.api.launch_preview`
  - `mission_control.api.command_submit`
- Metrics:
  - latency and error rate per route;
  - preview blocked ratio by `blocking_reason`;
  - realtime resync/invalidate rate.
- Traces:
  - HTTP -> gRPC -> repository path for snapshot and preview.

## Открытые вопросы
- Нужен ли отдельный query parameter для forced inclusion of recent closed context in list fallback, или bounded recent closed context остаётся строго automatic under the same fixed filters?

## Апрув
- request_id: `owner-2026-03-16-issue-519-api-contract`
- Решение: pending
- Комментарий: требуется owner review transport baseline и handover в `run:plan`.
