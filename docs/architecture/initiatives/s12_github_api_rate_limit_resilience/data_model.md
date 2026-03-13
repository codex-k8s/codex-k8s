---
doc_id: DM-S12-GITHUB-RL-0001
type: data-model
title: "GitHub API rate-limit resilience — Data model Sprint S12 Day 5"
status: approved
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-420-data-model"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Data Model: GitHub API rate-limit resilience

## TL;DR
- Schema owner остаётся `services/internal/control-plane`.
- Новые persisted сущности: `github_rate_limit_waits`, `github_rate_limit_wait_evidence`.
- `agent_runs` и `agent_sessions` расширяются как coarse pause engine, но source-of-truth для GitHub rate-limit semantics живёт в отдельном wait aggregate.
- Главный миграционный риск: корректный rollout новых wait enums/statuses и dominant wait election без drift между staff surfaces и resume sweeps.

## Сущности
### Entity: `github_rate_limit_waits`
- Назначение: canonical wait aggregate для recoverable GitHub rate-limit.
- Важные инварианты:
  - один open wait максимум на `(run_id, contour_kind)`;
  - один `dominant_for_run=true` максимум на run среди open waits;
  - hard failures не создают wait aggregate;
  - `resume_action_kind` и `resume_payload_json` задают единственный autoritative replay path.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | wait aggregate id |
| project_id | uuid | no |  | fk -> projects | tenant boundary |
| run_id | uuid | no |  | fk -> agent_runs | owning run |
| contour_kind | text | no |  | check(platform_pat/agent_bot_token) | user-facing contour |
| signal_origin | text | no |  | check(control_plane/worker/agent_runner) | who emitted the current signal |
| operation_class | text | no |  | check(run_status_comment/issue_label_transition/repository_provider_call/agent_github_call) | blocked activity |
| state | text | no | `open` | check(open/auto_resume_scheduled/auto_resume_in_progress/resolved/manual_action_required/cancelled) | aggregate state |
| limit_kind | text | no |  | check(primary/secondary) | provider classification |
| confidence | text | no | `deterministic` | check(deterministic/conservative/provider_uncertain) | user-facing confidence |
| recovery_hint_kind | text | no |  | check(rate_limit_reset/retry_after/exponential_backoff/manual_only) | recovery policy kind |
| dominant_for_run | bool | no | false |  | elected dominant wait |
| signal_id | text | no |  | unique | dedupe anchor for latest wait creation/update |
| request_fingerprint | text | yes |  |  | semantic fingerprint of blocked call |
| correlation_id | text | no |  | index | audit correlation |
| resume_action_kind | text | no |  | check(run_status_comment_retry/platform_github_call_replay/agent_session_resume) | replay path |
| resume_payload_json | jsonb | no | '{}'::jsonb |  | typed payload by `resume_action_kind` |
| manual_action_kind | text | yes |  | check(requeue_platform_operation/resume_agent_session/retry_after_operator_review) | terminal guidance |
| auto_resume_attempts_used | int | no | 0 |  | monotonic |
| max_auto_resume_attempts | int | no | 0 |  | policy cap |
| resume_not_before | timestamptz | yes |  | index | next safe retry lower bound |
| last_resume_attempt_at | timestamptz | yes |  |  | |
| first_detected_at | timestamptz | no | now() |  | initial wait entry |
| last_signal_at | timestamptz | no | now() |  | latest evidence time |
| resolved_at | timestamptz | yes |  |  | terminal timestamp |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: `github_rate_limit_wait_evidence`
- Назначение: append-only evidence ledger для detect/classify/resume/escalation lifecycle.
- Важные инварианты:
  - evidence rows are append-only;
  - raw auth headers, bearer tokens and full stderr are never persisted;
  - one `(wait_id, event_kind, signal_id)` tuple is unique when `signal_id` is present.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| wait_id | uuid | no |  | fk -> github_rate_limit_waits | owner wait |
| event_kind | text | no |  | check(signal_detected/classified/resume_scheduled/resume_attempted/resume_failed/resolved/manual_action_required/comment_mirror_failed) | lifecycle event |
| signal_id | text | yes |  |  | dedupe when event originated from reported signal |
| signal_origin | text | yes |  | check(control_plane/worker/agent_runner) | evidence source |
| provider_status_code | int | yes |  |  | `403`/`429` or null for synthetic lifecycle events |
| retry_after_seconds | int | yes |  |  | sanitized header |
| rate_limit_limit | int | yes |  |  | `x-ratelimit-limit` |
| rate_limit_remaining | int | yes |  |  | `x-ratelimit-remaining` |
| rate_limit_used | int | yes |  |  | `x-ratelimit-used` |
| rate_limit_reset_at | timestamptz | yes |  |  | `x-ratelimit-reset` converted to UTC |
| rate_limit_resource | text | yes |  |  | `x-ratelimit-resource` |
| github_request_id | text | yes |  |  | `x-github-request-id` |
| documentation_url | text | yes |  |  | provider docs/help link |
| message_excerpt | text | yes |  |  | normalized user-safe message |
| stderr_excerpt | text | yes |  |  | sanitized/truncated runner stderr excerpt |
| payload_json | jsonb | no | '{}'::jsonb |  | typed extras by `event_kind` |
| observed_at | timestamptz | no | now() | index | source event time |
| created_at | timestamptz | no | now() |  | |

### Entity: `agent_runs` (extension for GitHub rate-limit linkage)
- Назначение: coarse runtime state + dominant wait linkage.
- Важные инварианты:
  - `status=waiting_backpressure` always implies `wait_reason=github_rate_limit`;
  - `wait_target_kind=github_rate_limit_wait` and `wait_target_ref=<uuid>` point to the dominant open wait;
  - `wait_deadline_at` mirrors dominant `resume_not_before` only when deterministic/conservative retry exists.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| status | text | no |  | check(.../waiting_backpressure/...) | enum expands from existing run statuses |
| wait_reason | text | yes |  | check(owner_review/approval_pending/interaction_response/github_rate_limit) | closed business meaning |
| wait_target_kind | text | yes |  | check(approval_request/interaction_request/github_rate_limit_wait) | typed wait owner |
| wait_target_ref | text | yes |  |  | dominant wait id |
| wait_deadline_at | timestamptz | yes |  |  | mirrors `resume_not_before` when present |

### Entity: `agent_sessions` (extension for backpressure wait state)
- Назначение: session snapshot persistence and timeout guard for resumable runs.
- Day5 decision:
  - new columns are not required;
  - `wait_state` closed set expands with `backpressure`;
  - `timeout_guard_disabled=true` while GitHub wait is active.
- Причина:
  - existing session snapshot storage already supports deterministic resume;
  - no separate rate-limit session table is needed.

### Entity: `flow_events` (payload schema hardening)
- Назначение: audit-first traceability for visibility, manual guidance and comment mirror failures.
- New event payload keys:
  - `github_rate_limit.detected`
  - `github_rate_limit.wait.entered`
  - `github_rate_limit.resume.scheduled`
  - `github_rate_limit.resume.attempted`
  - `github_rate_limit.resume.succeeded`
  - `github_rate_limit.manual_action_required`
  - `github_rate_limit.visibility.comment_retry_scheduled`
- Ограничение:
  - full raw headers or token-bearing payloads are not copied into `flow_events.payload`;
  - only sanitized references, timestamps and IDs are mirrored.

## JSONB payload variant mapping
### `github_rate_limit_waits.resume_payload_json`
| `resume_action_kind` | Stored variant |
|---|---|
| `run_status_comment_retry` | `RunStatusCommentRetryPayload` |
| `platform_github_call_replay` | `PlatformGitHubCallReplayPayload` |
| `agent_session_resume` | `AgentSessionRateLimitResumePayload` |

### `github_rate_limit_wait_evidence.payload_json`
| `event_kind` | Stored variant |
|---|---|
| `signal_detected` | `GitHubRateLimitSignalPayload` |
| `classified` | `GitHubRateLimitClassificationPayload` |
| `resume_scheduled` | `GitHubRateLimitResumeSchedulePayload` |
| `resume_attempted` | `GitHubRateLimitResumeAttemptPayload` |
| `resume_failed` | `GitHubRateLimitResumeFailurePayload` |
| `manual_action_required` | `GitHubRateLimitManualActionPayload` |
| `comment_mirror_failed` | `GitHubRateLimitCommentMirrorPayload` |

## Связи
- `agent_runs 1:N github_rate_limit_waits`
- `github_rate_limit_waits 1:N github_rate_limit_wait_evidence`
- `agent_runs 1:1 dominant github_rate_limit_waits` via `wait_target_ref`
- `agent_runs 1:1 agent_sessions` (existing model, reused)

## Индексы и запросы (критичные)
- Partial unique open wait per contour:
  - `(run_id, contour_kind)` where `state in ('open','auto_resume_scheduled','auto_resume_in_progress','manual_action_required')`
- Dominant wait lookup:
  - partial unique `(run_id)` where `dominant_for_run=true and state in ('open','auto_resume_scheduled','auto_resume_in_progress','manual_action_required')`
- Resume sweeps:
  - `(state, resume_not_before)` filtered on `state in ('open','auto_resume_scheduled')`
- Evidence timeline:
  - `(wait_id, observed_at desc, id desc)`
- Wait queue:
  - `(project_id, state, dominant_for_run, updated_at desc)`

## Политика хранения данных
- Wait aggregates and evidence are retained for audit; no destructive cleanup policy is introduced on Day5.
- `stderr_excerpt` is truncated and sanitized before persistence.
- `documentation_url`, request ids and rate-limit header fields are safe to store; auth tokens are never stored.
- Manual-action guidance stays in persisted model even if feature flags are later disabled.

## Доменные инварианты
- `manual_action_required` is terminal for automatic retries: `auto_resume_attempts_used = max_auto_resume_attempts`.
- `resume_not_before` is nullable only for `manual_only` / provider-uncertain waits that already exhausted their auto-resume budget.
- `dominant_for_run=true` wait must always match `agent_runs.wait_target_ref`.
- `state=resolved` requires `resolved_at`.
- `agent_session_resume` payload is produced only after a successful terminal resolution and before runner resume is triggered.

## Ownership and write path
- `control-plane`:
  - owns schema, classification, dominant wait election, visibility projection and comment render context.
- `worker`:
  - writes only through owner use-cases for resume sweeps and manual escalation.
- `agent-runner`:
  - never writes wait rows directly; it only reports signals and persists session snapshots through existing callbacks.

## Continuity after `run:plan`
- Plan package Issue `#423` закрепил этот data model как baseline для waves `#425`, `#426`, `#427` и `#428`.
- Execution streams не могут добавлять parallel source-of-truth для wait semantics вне `github_rate_limit_waits`, `github_rate_limit_wait_evidence` и зафиксированных расширений `agent_runs`/`agent_sessions`.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): absent, docs only.
- Migration impact (`run:dev`):
  - additive creation of wait/evidence tables;
  - enum/check expansion for `agent_runs` and `agent_sessions`;
  - no shared-DB ownership change.

## Context7 validation
- GitHub rate-limit semantics re-checked through Context7 `/github/docs`.
- New external dependencies for data-model part are not required.
