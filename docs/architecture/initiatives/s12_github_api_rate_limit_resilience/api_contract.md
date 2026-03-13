---
doc_id: API-S12-GITHUB-RL-0001
type: api-contract
title: "GitHub API rate-limit resilience — API contract Sprint S12 Day 5"
status: in-review
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-issue-420-api-contract"
---

# API Contract: GitHub API rate-limit resilience

## TL;DR
- Контрактный scope: internal `agent-runner -> control-plane` signal handoff, existing staff/private run visibility surfaces, realtime updates and GitHub service-comment render context.
- Аутентификация: run-bound bearer token for internal callback RPC, staff JWT for visibility surfaces, platform GitHub credentials only inside domain adapters.
- Версионирование: existing `/api/v1/staff/...` routes are extended additively; new internal gRPC callback is versioned in `controlplane.v1`.
- Общий принцип: edge stays thin; `control-plane` alone owns classification, wait projection and render semantics.

## Спецификации (source of truth)
- Future OpenAPI source of truth: `services/external/api-gateway/api/server/api.yaml`
- Future gRPC source of truth: `proto/codexk8s/controlplane/v1/controlplane.proto`
- Design-stage interim sources:
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/design_doc.md`
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/api_contract.md`

## Operations / Methods
| Operation | Method/Kind | Path/Name | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Report agent-side rate-limit signal | gRPC | `ReportGitHubRateLimitSignal` | run-bound bearer | `signal_id` | `agent-runner -> control-plane`; runner stops local retries after success |
| Read run details with wait projection | HTTP GET | `/api/v1/staff/runs/{run_id}` | staff JWT | n/a | existing route extended with `wait_projection` |
| List active waits | HTTP GET | `/api/v1/staff/runs/waits` | staff JWT | n/a | existing route extended with contour and wait policy fields |
| Stream run realtime wait updates | WebSocket | `/api/v1/staff/runs/{run_id}/realtime` | staff JWT | n/a | existing route adds `wait_entered|wait_updated|wait_resolved|wait_manual_action_required` envelopes |
| Render GitHub status mirror | internal domain op | `UpsertRunStatusComment` | platform contour | `correlation_id` | best-effort mirror derived from wait projection |

## Internal callback contract (`agent-runner -> control-plane`)
### `ReportGitHubRateLimitSignalRequest`
| Field | Type | Required | Notes |
|---|---|---|---|
| `run_id` | uuid | yes | owning run |
| `signal_id` | string | yes | stable dedupe key for this detection |
| `correlation_id` | string | yes | audit correlation |
| `contour_kind` | `agent_bot_token` | yes | fixed for runner path |
| `signal_origin` | `agent_runner` | yes | fixed for this RPC |
| `operation_class` | `agent_github_call` | yes | current Day5 scope |
| `provider_status_code` | int32 | yes | expected `403` or `429` |
| `occurred_at` | RFC3339 timestamp | yes | detection time in UTC |
| `request_fingerprint` | string | no | stable fingerprint of blocked call |
| `stderr_excerpt` | string | no | sanitized excerpt, max 4 KiB |
| `message_excerpt` | string | no | normalized user-safe excerpt |
| `github_headers` | `GitHubRateLimitHeaders` | no | typed header snapshot |
| `session_snapshot_version` | int64 | no | latest persisted snapshot version from runner |

### `GitHubRateLimitHeaders`
| Field | Type | Notes |
|---|---|---|
| `rate_limit_limit` | int32 | mirrors `x-ratelimit-limit` |
| `rate_limit_remaining` | int32 | mirrors `x-ratelimit-remaining` |
| `rate_limit_used` | int32 | mirrors `x-ratelimit-used` |
| `rate_limit_reset_at` | RFC3339 timestamp | converted from `x-ratelimit-reset` |
| `rate_limit_resource` | string | mirrors `x-ratelimit-resource` |
| `retry_after_seconds` | int32 | mirrors `retry-after` when present |
| `github_request_id` | string | `x-github-request-id` |
| `documentation_url` | string | provider help URL when present |

### `ReportGitHubRateLimitSignalResponse`
| Field | Type | Notes |
|---|---|---|
| `wait_id` | uuid | created or reused dominant wait |
| `wait_state` | `waiting_backpressure` | coarse runtime state |
| `wait_reason` | `github_rate_limit` | business meaning |
| `next_step_kind` | `auto_resume_scheduled|manual_action_required` | current resolution path |
| `runner_action` | `persist_session_and_exit_wait` | runner must stop local retries |
| `resume_not_before` | RFC3339 timestamp | set for deterministic/conservative waits only |

## Staff/private visibility contract
### `RunWaitProjection`
| Field | Type | Notes |
|---|---|---|
| `wait_state` | `waiting_backpressure` | coarse runtime state |
| `wait_reason` | `github_rate_limit` | domain reason |
| `dominant_wait` | `GitHubRateLimitWaitItem` | primary wait for runtime semantics |
| `related_waits` | `GitHubRateLimitWaitItem[]` | additional open contour waits |
| `comment_mirror_state` | `synced|pending_retry|not_attempted` | GitHub comment mirror health |

### `GitHubRateLimitWaitItem`
| Field | Type | Notes |
|---|---|---|
| `wait_id` | uuid | aggregate id |
| `contour_kind` | `platform_pat|agent_bot_token` | user-facing operational contour |
| `limit_kind` | `primary|secondary` | provider signal kind |
| `operation_class` | `run_status_comment|issue_label_transition|repository_provider_call|agent_github_call` | blocked activity |
| `state` | `open|auto_resume_scheduled|auto_resume_in_progress|resolved|manual_action_required` | aggregate state |
| `confidence` | `deterministic|conservative|provider_uncertain` | user-facing confidence |
| `entered_at` | RFC3339 timestamp | wait start |
| `resume_not_before` | RFC3339 timestamp, optional | earliest safe retry |
| `attempts_used` | int32 | auto-resume attempts already spent |
| `max_attempts` | int32 | finite attempt budget |
| `recovery_hint` | `GitHubRateLimitRecoveryHint` | typed hint |
| `manual_action` | `GitHubRateLimitManualAction`, optional | present only when manual action required |

### `GitHubRateLimitRecoveryHint`
| Field | Type | Notes |
|---|---|---|
| `hint_kind` | `rate_limit_reset|retry_after|exponential_backoff|manual_only` | closed enum |
| `resume_not_before` | RFC3339 timestamp, optional | repeated for clients that only read hint |
| `source_headers` | `reset_at|retry_after|provider_uncertain` | provenance of hint |
| `details_markdown` | string | user-safe explanation |

### `GitHubRateLimitManualAction`
| Field | Type | Notes |
|---|---|---|
| `kind` | `requeue_platform_operation|resume_agent_session|retry_after_operator_review` | closed enum |
| `summary` | string | short CTA |
| `details_markdown` | string | audit-safe operator guidance |
| `suggested_not_before` | RFC3339 timestamp, optional | conservative lower bound for retry |

## Realtime extension
### `RunRealtimeWaitEnvelope`
| `event_kind` | Payload | Required fields |
|---|---|---|
| `wait_entered` | `RunWaitProjection` | `wait_state`, `dominant_wait`, `comment_mirror_state` |
| `wait_updated` | `RunWaitProjection` | same as above |
| `wait_resolved` | `RunWaitResolution` | `wait_id`, `resolution_kind`, `resolved_at` |
| `wait_manual_action_required` | `RunWaitManualActionEvent` | `wait_id`, `manual_action`, `updated_at` |

### `RunWaitResolution`
| Field | Type | Notes |
|---|---|---|
| `wait_id` | uuid | resolved aggregate |
| `contour_kind` | enum | resolved contour |
| `resolution_kind` | `auto_resumed|manually_resolved|cancelled` | terminal outcome |
| `resolved_at` | RFC3339 timestamp | terminal time |

## GitHub service-comment render context
- Service-comment mirror uses a dedicated internal typed model:
  - `headline`
  - `dominant_contour`
  - `limit_kind`
  - `operation_class`
  - `next_step_kind`
  - `resume_not_before`
  - `manual_action_summary`
  - `related_contour_badges[]`
- Normative rules:
  - render context is produced only from persisted wait projection;
  - raw stderr and raw headers never reach the comment template;
  - if platform contour blocks comment sync, system persists `comment_mirror_state=pending_retry` instead of fabricating success.

## Error model
- Canonical domain codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- `ReportGitHubRateLimitSignal` mapping:
  - malformed headers or missing `signal_id` -> `invalid_argument`
  - invalid run token -> `unauthorized`
  - unknown run -> `not_found`
  - duplicate stale signal that loses dominance election -> `conflict`
  - run not resumable / already terminal -> `failed_precondition`
  - persistence failure -> `internal`
- Visibility routes:
  - wait projection missing for non-wait run is not an error; field stays absent.

## Retries / rate limits
- `ReportGitHubRateLimitSignal` is idempotent by `signal_id`.
- Auto-resume replays are serialized per `wait_id`; the system never fans out concurrent replays for one wait.
- Platform mutative GitHub retries must be spaced by at least 1 second between attempts, matching GitHub best-practice guidance.
- `GET /rate_limit` is explicitly not part of the recovery loop:
  - GitHub docs say response headers should be preferred;
  - the endpoint may still count against secondary rate limit.

## Контракты данных (DTO)
- Closed enums:
  - `contour_kind`
  - `limit_kind`
  - `operation_class`
  - `hint_kind`
  - `manual_action_kind`
  - `next_step_kind`
- Запрещено:
  - `map[string]any` / `any` in signal or wait projection DTO;
  - free-form `operation_class` strings;
  - embedding raw auth headers or tokens in any DTO.

## Backward compatibility
- Project is pre-production, so coordinated breaking changes are allowed.
- Day5 rollout still keeps additive discipline:
  - existing run detail / wait queue routes are extended, not replaced;
  - worker and runner must be deployed only after `control-plane` understands new wait projection;
  - UI may temporarily ignore `wait_projection` until `CODEXK8S_GITHUB_RATE_LIMIT_WAIT_UI_ENABLED=true`.

## Наблюдаемость
- Logs:
  - `github_rate_limit.signal_callback.accepted`
  - `github_rate_limit.wait_projection.read`
  - `github_rate_limit.comment_rendered`
  - `github_rate_limit.comment_retry_pending`
- Metrics:
  - `staff_run_wait_projection_total{state}`
  - `github_rate_limit_signal_callback_total{result}`
  - `github_rate_limit_comment_render_total{result}`
- Traces:
  - `agent-runner.callback -> control-plane.persist`
  - `staff-http -> control-plane.wait-projection`

## Context7 validation
- Через Context7 `/github/docs` подтверждены:
  - primary/secondary rate-limit semantics;
  - приоритет response headers over `GET /rate_limit`;
  - guidance `wait at least one minute` and `exponential backoff` when `Retry-After` отсутствует;
  - avoidance of concurrency bursts.
- Новые внешние библиотеки Day5 не выбирались.
