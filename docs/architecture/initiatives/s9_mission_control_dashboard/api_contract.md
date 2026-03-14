---
doc_id: API-S9-MISSION-CONTROL-0001
type: api-contract
title: "Mission Control Dashboard — API contract Sprint S9 Day 5"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-14
related_issues: [333, 335, 337, 340, 351, 363]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-351-api-contract"
---

# API Contract: Mission Control Dashboard

## TL;DR
- Тип API: staff/private REST + internal gRPC + WebSocket realtime stream.
- Аутентификация: staff JWT + project RBAC; mutating operations требуют role/policy check в `control-plane`.
- Версионирование: `/api/v1` для HTTP и `v1` package для gRPC/realtime envelope version.
- Основные операции: dashboard snapshot, entity details, command submit/status, realtime updates, optional voice candidate draft/promotion.

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
| Get dashboard snapshot | GET | `/api/v1/staff/mission-control/dashboard` | staff JWT | n/a | Query: `view_mode`, `active_filter`, `search`, `cursor`, `limit` |
| Get entity details | GET | `/api/v1/staff/mission-control/entities/{entity_kind}/{entity_public_id}` | staff JWT | n/a | Returns detail payload, relations, timeline preview, allowed actions |
| List entity timeline | GET | `/api/v1/staff/mission-control/entities/{entity_kind}/{entity_public_id}/timeline` | staff JWT | n/a | Cursor pagination, read-only provider comments in MVP |
| Submit command | POST | `/api/v1/staff/mission-control/commands` | staff JWT | `Idempotency-Key` | Admits typed command and returns command acknowledgement |
| Get command status | GET | `/api/v1/staff/mission-control/commands/{command_id}` | staff JWT | n/a | Poll/read fallback when realtime unavailable |
| Connect realtime stream | GET | `/api/v1/staff/mission-control/realtime` | staff JWT | n/a | WebSocket upgrade; requires `resume_token` from latest snapshot |
| Create voice candidate | POST | `/api/v1/staff/mission-control/voice-candidates` | staff JWT + policy | `Idempotency-Key` | Optional; only when `CODEXK8S_MISSION_CONTROL_VOICE_ENABLED=true` |
| Promote voice candidate | POST | `/api/v1/staff/mission-control/voice-candidates/{candidate_id}/promote` | staff JWT + policy | `Idempotency-Key` | Creates linked discussion/task or command |
| Reject voice candidate | POST | `/api/v1/staff/mission-control/voice-candidates/{candidate_id}/reject` | staff JWT + policy | `Idempotency-Key` | Marks candidate as rejected, no provider mutation |

## Query semantics
- `view_mode`:
  - `board`
  - `list`
- `active_filter`:
  - `working`
  - `waiting`
  - `blocked`
  - `review`
  - `recent_critical_updates`
  - `all_active`
- `cursor`:
  - opaque token for pagination within the same snapshot scope.
- `search`:
  - substring search over typed indexed fields and projection payload summaries; never arbitrary client-side filtering only.

## Internal gRPC methods (design baseline)
| RPC | Request | Response | Error mapping |
|---|---|---|---|
| `GetMissionControlSnapshot` | `GetMissionControlSnapshotRequest` | `GetMissionControlSnapshotResponse` | `invalid_argument`, `forbidden`, `internal` |
| `GetMissionControlEntity` | `GetMissionControlEntityRequest` | `GetMissionControlEntityResponse` | `not_found`, `forbidden` |
| `ListMissionControlTimeline` | `ListMissionControlTimelineRequest` | `ListMissionControlTimelineResponse` | `not_found`, `forbidden` |
| `SubmitMissionControlCommand` | `SubmitMissionControlCommandRequest` | `SubmitMissionControlCommandResponse` | `invalid_argument`, `conflict`, `failed_precondition`, `forbidden` |
| `GetMissionControlCommand` | `GetMissionControlCommandRequest` | `GetMissionControlCommandResponse` | `not_found`, `forbidden` |
| `OpenMissionControlRealtime` | `OpenMissionControlRealtimeRequest` | stream `MissionControlRealtimeEnvelope` | `failed_precondition`, `unauthorized` |
| `CreateVoiceCandidate` | `CreateVoiceCandidateRequest` | `CreateVoiceCandidateResponse` | `failed_precondition`, `forbidden` |
| `PromoteVoiceCandidate` | `PromoteVoiceCandidateRequest` | `PromoteVoiceCandidateResponse` | `conflict`, `failed_precondition`, `forbidden` |
| `RejectVoiceCandidate` | `RejectVoiceCandidateRequest` | `RejectVoiceCandidateResponse` | `failed_precondition`, `forbidden` |

## Public identifier contract (normative)
- `entity_public_id` is the only stable public entity identifier in HTTP/gRPC DTO and route params.
- `entity_public_id` maps 1:1 to `mission_control_entities.entity_external_key`.
- `mission_control_entities.id` remains an internal persistence key for joins/indexes and MUST NOT leak into transport DTO, deep links or relation refs.
- Any entity reference in transport is always the pair `(entity_kind, entity_public_id)`, never a raw bigint row id.

## DTO contract (key models)
### DashboardSnapshot DTO
- `snapshot_id` (string, opaque)
- `view_mode` (`board|list`)
- `freshness_status` (`fresh|stale|degraded`)
- `generated_at` (RFC3339)
- `stale_after` (RFC3339)
- `realtime_resume_token` (string)
- `summary`:
  - `total_entities`
  - `working_count`
  - `waiting_count`
  - `blocked_count`
  - `review_count`
  - `recent_critical_updates_count`
- `entities[]` (`MissionControlEntityCard`)
- `relations[]` (`MissionControlRelation`)
- `next_page_cursor` (string, optional)

### MissionControlEntityCard DTO
- `entity_kind` (`work_item|discussion|pull_request|agent`)
- `entity_public_id` (string, stable public identifier)
- `title` (string)
- `state` (`working|waiting|blocked|review|recent_critical_updates`)
- `sync_status` (`synced|pending_sync|failed|degraded`)
- `provider_reference` (`provider`, `external_id`, `url`)
- `primary_actor` (`actor_type`, `actor_id`, `display_name`)
- `relation_count` (int32)
- `last_timeline_at` (RFC3339)
- `badges[]` (`blocked`, `owner_review`, `waiting_mcp`, `realtime_stale`, `voice_candidate`)

### MissionControlEntityDetails DTO
- `entity` (`MissionControlEntityCard`)
- `detail_payload` (typed object per `entity_kind`)
- `relations[]`
- `timeline_preview[]`
- `allowed_actions[]`
- `provider_deep_links[]`

### MissionControlRelation DTO
- `relation_kind` (`linked_to|blocks|blocked_by|formalized_from|owned_by|assigned_to|tracked_by_command`)
- `source_entity_kind`
- `source_entity_public_id`
- `target_entity_kind`
- `target_entity_public_id`
- `direction` (`outbound|inbound|bidirectional`)

### MissionControlTimelineEntry DTO
- `entry_id` (string)
- `entity_kind`
- `entity_public_id`
- `source_kind` (`provider|platform|command|voice_candidate`)
- `source_ref` (string)
- `occurred_at` (RFC3339)
- `summary` (string)
- `body_markdown` (string, optional)
- `command_id` (string, optional)
- `provider_url` (string, optional)
- `is_read_only` (bool)

### MissionControlAllowedAction DTO
- `action_kind` (`discussion.create|work_item.create|discussion.formalize|stage.next_step.execute|command.retry_sync`)
- `presentation` (`primary|secondary|danger|link`)
- `allowed_when_degraded` (bool)
- `approval_requirement` (`none|owner_review`)
- `blocked_reason` (string, optional)
- `command_template` (typed template object, optional)

### MissionControlProviderDeepLink DTO
- `action_kind` (`provider.open_issue|provider.open_pr|provider.review|provider.merge|provider.reply_comment`)
- `url`
- `reason` (`provider_policy`, `not_in_mvp_inline_scope`, `requires_human_review`, `unsafe_when_degraded`)

### MissionControlCommandRequest DTO
- `command_kind`:
  - `discussion.create`
  - `work_item.create`
  - `discussion.formalize`
  - `stage.next_step.execute`
  - `command.retry_sync`
- `project_id` (uuid)
- `target_entity_kind` (optional for create commands)
- `target_entity_public_id` (optional for create commands)
- `business_intent_key` (string, required)
- `expected_projection_version` (int64, optional but required for commands against existing entity)
- `payload` (typed union per `command_kind`)

### MissionControlCommandAck DTO
- `command_id` (string)
- `command_kind`
- `status` (`accepted|pending_approval|queued|pending_sync|blocked`)
- `business_intent_key`
- `correlation_id`
- `acknowledged_at` (RFC3339)
- `entity_refs[]`
- `approval` (`approval_state`, `approval_request_id`, `requested_at`, optional)
- `blocking_reason` (optional typed enum)

### MissionControlCommandStatus DTO
- `command_id`
- `status` (`accepted|pending_approval|queued|pending_sync|reconciled|failed|blocked|cancelled`)
- `failure_reason` (`provider_error|policy_denied|projection_stale|duplicate_intent|timeout|approval_denied|approval_expired|unknown`, optional)
- `status_message` (string)
- `updated_at` (RFC3339)
- `entity_refs[]`
- `provider_delivery_ids[]`
- `approval` (`approval_state`, `approval_request_id`, `requested_at`, `decided_at`, `approver_actor_id`, optional)

### MissionControlRealtimeEnvelope DTO
- `event_kind` (`connected|delta|invalidate|stale|degraded|resync_required|heartbeat|error`)
- `snapshot_id`
- `resume_token`
- `occurred_at` (RFC3339)
- `payload` (typed union per `event_kind`)

### VoiceCandidate DTO
- `candidate_id` (string)
- `status` (`draft|promoted|rejected|expired`)
- `source_kind` (`voice|audio_upload|transcript`)
- `transcript_excerpt` (string)
- `structured_summary` (string)
- `confidence` (float, optional)
- `linked_entity_kind` (optional)
- `linked_entity_public_id` (optional)

### `detail_payload` variants for `MissionControlEntityDetails`
| `entity_kind` | DTO variant | Required fields |
|---|---|---|
| `work_item` | `WorkItemDetailsPayload` | `work_item_type`, `stage_label`, `labels[]`, `owner`, `assignees[]`, `last_provider_sync_at` |
| `discussion` | `DiscussionDetailsPayload` | `discussion_kind`, `status`, `author`, `participant_count`, `latest_comment_excerpt`, `formalization_target` |
| `pull_request` | `PullRequestDetailsPayload` | `branch_head`, `branch_base`, `merge_state`, `review_decision`, `checks_summary`, `linked_issue_refs[]` |
| `agent` | `AgentDetailsPayload` | `agent_key`, `run_status`, `runtime_mode`, `waiting_reason`, `active_run_id`, `last_heartbeat_at` |

### `command_template` variants for `MissionControlAllowedAction`
| `action_kind` | DTO variant | Required fields |
|---|---|---|
| `discussion.create` | `DiscussionCreateTemplate` | `default_title`, `default_body_markdown`, `parent_entity_kind`, `parent_entity_public_id` |
| `work_item.create` | `WorkItemCreateTemplate` | `suggested_title`, `suggested_body_markdown`, `default_labels[]`, `default_assignees[]` |
| `discussion.formalize` | `DiscussionFormalizeTemplate` | `source_entity_kind`, `source_entity_public_id`, `target_kind`, `suggested_title`, `suggested_body_markdown` |
| `stage.next_step.execute` | `StageNextStepTemplate` | `thread_kind`, `thread_number`, `target_label`, `removed_labels[]`, `display_variant`, `approval_requirement` |
| `command.retry_sync` | `RetrySyncTemplate` | `command_id`, `latest_failure_reason`, `can_retry_after`, `diagnostic_hint` |

### `payload` variants for `MissionControlCommandRequest`
| `command_kind` | DTO variant | Required fields |
|---|---|---|
| `discussion.create` | `DiscussionCreatePayload` | `title`, `body_markdown`, `parent_entity_kind`, `parent_entity_public_id` |
| `work_item.create` | `WorkItemCreatePayload` | `title`, `body_markdown`, `initial_labels[]`, `related_entity_refs[]` |
| `discussion.formalize` | `DiscussionFormalizePayload` | `source_entity_kind`, `source_entity_public_id`, `formalized_kind`, `title`, `body_markdown` |
| `stage.next_step.execute` | `StageNextStepExecutePayload` | `thread_kind`, `thread_number`, `target_label`, `removed_labels[]`, `display_variant`, `approval_requirement` |
| `command.retry_sync` | `RetrySyncPayload` | `command_id`, `retry_reason`, `expected_status` |

### `payload` variants for `MissionControlRealtimeEnvelope`
| `event_kind` | DTO variant | Required fields |
|---|---|---|
| `connected` | `ConnectedRealtimePayload` | `snapshot_freshness_status`, `server_cursor` |
| `delta` | `DeltaRealtimePayload` | `delta_entities[]`, `delta_relations[]`, `delta_timeline_entries[]`, `changed_command_ids[]` |
| `invalidate` | `InvalidateRealtimePayload` | `reason`, `refresh_scope`, `affected_entity_refs[]` |
| `stale` | `StaleRealtimePayload` | `reason`, `stale_since`, `suggested_refresh` |
| `degraded` | `DegradedRealtimePayload` | `reason`, `fallback_mode`, `affected_capabilities[]` |
| `resync_required` | `ResyncRequiredRealtimePayload` | `reason`, `required_snapshot_id`, `dropped_event_count` |
| `heartbeat` | `HeartbeatRealtimePayload` | `server_time`, `snapshot_id` |
| `error` | `ErrorRealtimePayload` | `code`, `message`, `retryable` |

## Validation and guardrails
- Contract rules:
  - `entity_kind` is closed enum, no free-form resource names.
  - `entity_public_id` is the exact transport mirror of `entity_external_key`; internal bigint row ids are never accepted or returned.
  - `command_kind` is closed enum, unknown commands rejected with `invalid_argument`.
  - `business_intent_key` required for every mutating command and voice promotion.
  - `expected_projection_version` required when mutating an existing entity to prevent acting on stale client context.
  - `stage.next_step.execute` may return `pending_approval`; while in this state provider mutations and label transitions MUST NOT be attempted yet.
  - `blocked` denotes a hard stop or a denied/expired approval decision, not "still waiting for owner review".
- Deep-link rules:
  - provider deep links are read-only transport objects, not pseudo-command fallbacks.
- Realtime rules:
  - `resume_token` expires with snapshot freshness window;
  - on `resync_required`, client must fetch a new snapshot before applying more deltas.

## Error model
- Canonical domain codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- Specific mappings:
  - stale projection or invalid `expected_projection_version` -> `failed_precondition`
  - duplicate `business_intent_key` with same semantic payload -> `conflict` + current command reference
  - command not allowed in degraded mode -> `failed_precondition`
  - voice disabled by policy -> `failed_precondition`

## Retries / rate limits
- Safe retries:
  - POST commands and voice candidate mutations only with `Idempotency-Key`
  - GET snapshot/details/command status/timeline
- Rate limits baseline:
  - snapshot/details protected per user/project
  - realtime connections limited per user/session
  - voice candidate create/promote limited stricter than core dashboard reads

## Backward compatibility
- Initiative remains pre-production; coordinated breaking changes inside staff/private API are acceptable.
- Rollout discipline remains strict:
  - DB schema first
  - internal gRPC/domain second
  - edge transport third
  - frontend last
- If realtime endpoint is unavailable, HTTP snapshot/details/command status remain canonical fallback.

## Provider deep-link-only scope (normative)
- The following provider actions MUST stay out of inline staff write-path in MVP:
  - PR review submit
  - PR merge/rebase
  - comment reply/edit/delete
  - raw label edits in provider UI
  - direct assignee/reviewer mutations without platform-safe command contract
- These actions MAY be surfaced only as `provider_deep_links[]`.

## Наблюдаемость
- Logs:
  - endpoint/rpc, `project_id`, `entity_kind`, `entity_public_id`, `command_kind`, `status`, `correlation_id`, `duration_ms`
- Metrics:
  - `mission_control_http_requests_total{endpoint,code}`
  - `mission_control_command_total{kind,status}`
  - `mission_control_realtime_connections`
  - `mission_control_realtime_resync_total`
  - `mission_control_voice_candidate_total{status}`
- Traces:
  - snapshot/details HTTP spans
  - command admission spans
  - realtime publish spans

## Context7 validation
- Через Context7 подтверждён актуальный CLI syntax для `gh issue create`, `gh pr create`, `gh pr edit`:
  - `/websites/cli_github_manual`
- Новые внешние runtime библиотеки этим design-этапом не выбираются.
