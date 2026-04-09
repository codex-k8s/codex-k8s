---
doc_id: API-S17-CK8S-0001
type: api-contract
title: "Sprint S17 Day 5 — API contract for unified owner feedback loop (Issue #568)"
status: in-review
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [541, 554, 557, 559, 568, 575]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-568-api-contract"
---

# API Contract: Sprint S17 unified owner feedback loop

## TL;DR
- Контрактный scope: built-in wait path на `user.decision.request`, Telegram delivery/callback path поверх S11 adapter baseline и новый typed staff-console fallback contract.
- Аутентификация: run-bound bearer для MCP tool и resume lookup; interaction-scoped callback bearer для Telegram adapter; staff JWT для staff read/write endpoints; internal gRPC auth between `api-gateway` and `control-plane`.
- Версионирование: staff/private API остаётся на `/api/v1/...`; Telegram adapter delivery/callback payload stays on adapter path and uses `telegram-owner-feedback-v1`; internal gRPC bridge remains in `control-plane`.
- Общий принцип: channel semantics stay outside agent-visible tool input, while winner selection and lifecycle classification stay exclusively in `control-plane`.

## Спецификации (source of truth)
- Future OpenAPI source of truth:
  - `services/external/api-gateway/api/server/api.yaml`
  - `services/external/telegram-interaction-adapter/api/server/api.yaml`
- Future gRPC source of truth:
  - `proto/kodex/controlplane/v1/controlplane.proto`
- Design-stage interim source:
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/design_doc.md`
  - `docs/architecture/initiatives/s17_unified_owner_feedback_loop/api_contract.md`
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/api_contract.md`

## Operations / Methods
| Operation | Method/Kind | Path/Name | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Request owner feedback wait | MCP tool | `user.decision.request` | run-bound bearer | `(run_id, mcp_request_id)` | primary live wait entrypoint |
| List owner feedback requests | HTTP GET | `/api/v1/staff/owner-feedback/requests` | staff JWT | read-only | pending/terminal projections |
| Get owner feedback request | HTTP GET | `/api/v1/staff/owner-feedback/requests/{request_id}` | staff JWT | read-only | detail with bindings and allowed actions |
| Submit fallback response | HTTP POST | `/api/v1/staff/owner-feedback/requests/{request_id}/responses` | staff JWT | `(request_id, submission_id)` | staff-console option/free-text response |
| Telegram delivery | HTTP POST | adapter-side `/v1/telegram/interaction-deliveries` | adapter credential | `delivery_id` | worker -> adapter contour |
| Telegram callback | HTTP POST | `/api/v1/mcp/interactions/callback` | callback bearer token | `(interaction_id, adapter_event_id)` | adapter -> `api-gateway` |
| Submit callback | gRPC | `SubmitInteractionCallback` | internal service auth | `(interaction_id, adapter_event_id)` | `api-gateway` -> `control-plane` |
| Resume lookup | gRPC | `GetRunInteractionResumePayload` | run-bound bearer | `run_id` | `agent-runner` fetches typed result |

## Agent-facing wait contract
### `user.decision.request`
#### Input DTO
- S17 keeps the S10/S11 input surface unchanged:
  - `question`
  - `details_markdown`
  - `options`
  - `allow_free_text`
  - `free_text_placeholder`
  - `response_ttl_seconds`
- Additional S17 rules:
  - no recipient/channel fields in tool input;
  - no request to force `staff-console only`;
  - `response_ttl_seconds` must not imply effective wait shorter than owner wait window policy.

#### Initial output DTO
| Field | Type | Notes |
|---|---|---|
| `status` | `pending_user_response` | S10-compatible async ack |
| `interaction_id` | uuid | canonical request id |
| `wait_state` | `waiting_mcp` | coarse runtime state |
| `wait_reason` | `interaction_response` | business meaning |
| `expires_at` | RFC3339 timestamp | hard owner response deadline |
| `continuation_mode` | `live_same_session` | primary continuation contract |
| `same_session_required` | bool | always `true` for S17 owner feedback waits |
| `response_surfaces` | `ResponseSurface[]` | Telegram primary + staff fallback |
| `correlation_id` | string | platform correlation for audit / projections |

#### `ResponseSurface`
| Field | Type | Notes |
|---|---|---|
| `surface_kind` | `telegram_primary|staff_console_fallback` | |
| `availability` | `available|degraded` | initial platform view |
| `supports_option_response` | bool | |
| `supports_free_text_response` | bool | |

#### Resume payload DTO
| Field | Type | Notes |
|---|---|---|
| `interaction_id` | uuid | same aggregate id |
| `tool_name` | string | always `user.decision.request` |
| `request_status` | `answered|expired|manual_fallback|recovery_resumed` | terminal owner-feedback outcome |
| `response_kind` | `option|free_text|none` | `none` for expired without accepted response |
| `selected_option_id` | string | set only for option response |
| `free_text` | string | set only for free-text/voice transcription |
| `response_source_kind` | `telegram_callback|telegram_free_text|telegram_voice|staff_option|staff_free_text|none` | final accepted source |
| `continuation_path` | `live_same_session|recovery_resume|manual_fallback|expired` | typed continuation classification |
| `resolved_at` | RFC3339 timestamp | terminal timestamp |
| `resolution_reason` | `accepted|expired|manual_fallback|recovery_only` | audit-safe final classification |

#### Validation rules
- open owner-feedback wait on the same run -> `failed_precondition`
- duplicate `option_id` or invalid TTL -> `invalid_argument`
- any attempt to route to `owner.feedback.request` semantics or approval-only status vocabulary -> `failed_precondition`
- if effective wait timeout/TTL is below owner wait window policy, tool call is rejected as `failed_precondition`

## Staff-console fallback contract
### List response DTO `OwnerFeedbackRequestListItem`
| Field | Type | Notes |
|---|---|---|
| `request_id` | uuid | canonical request id |
| `run_id` | uuid | owner run |
| `issue_number` | int | issue linkage |
| `question_summary` | string | concise pending prompt |
| `canonical_status` | `delivery_pending|delivery_accepted|waiting|overdue|expired|manual_fallback|continuation_live|recovery_resume|resolved` | platform truth |
| `primary_surface_state` | `nominal|degraded|manual_fallback` | Telegram delivery posture |
| `owner_wait_deadline_at` | RFC3339 timestamp | hard deadline |
| `overdue_at` | RFC3339 timestamp | soft visibility threshold |
| `continuation_path` | `live_same_session|recovery_resume|manual_fallback|expired` | current path |
| `allowed_actions` | `AllowedAction[]` | typed action matrix |

### Detail response DTO `OwnerFeedbackRequestDetail`
| Field | Type | Notes |
|---|---|---|
| `request_id` | uuid | |
| `question` | string | full question |
| `details_markdown` | string | optional details |
| `options` | `OwnerFeedbackOptionBinding[]` | option bindings for staff console |
| `free_text_binding_id` | uuid | optional binding for free-text fallback |
| `canonical_status` | string | same as list |
| `response_deadline_at` | RFC3339 timestamp | |
| `projection_updated_at` | RFC3339 timestamp | staleness visibility |
| `allowed_actions` | `AllowedAction[]` | domain-derived, not UI-local |
| `last_visibility_signal` | `VisibilitySignal` | current reason for overdue/manual fallback/recovery |

### `VisibilitySignal`
| Field | Type | Notes |
|---|---|---|
| `signal_kind` | `none|overdue|expired|manual_fallback|recovery_resume` | current visible reason |
| `raised_at` | RFC3339 timestamp | when signal became active |
| `details` | string | diagnostic-safe explanation |

### `OwnerFeedbackOptionBinding`
| Field | Type | Notes |
|---|---|---|
| `binding_id` | uuid | opaque staff action id |
| `option_id` | string | semantic option id |
| `label` | string | owner-facing label |
| `description` | string | optional |
| `expires_at` | RFC3339 timestamp | bound to request deadline |

### `AllowedAction`
| Field | Type | Notes |
|---|---|---|
| `action_kind` | `respond_option|respond_free_text|view_only` | only values allowed in Sprint S17 |
| `enabled` | bool | |
| `reason` | string | diagnostic-safe explanation |

### Submit response request DTO `OwnerFeedbackResponseCommand`
| Field | Type | Required | Notes |
|---|---|---|---|
| `submission_id` | uuid | yes | idempotency key |
| `response_kind` | `option|free_text` | yes | |
| `binding_id` | uuid | yes | one of allowed bindings from detail DTO |
| `selected_option_id` | string | no | required for `option` |
| `free_text` | string | no | required for `free_text`; max `8192` UTF-8 bytes |

### Submit response outcome DTO `OwnerFeedbackResponseOutcome`
| Field | Type | Notes |
|---|---|---|
| `accepted` | bool | whether request state changed |
| `classification` | `applied|duplicate|stale|expired|invalid` | deterministic result |
| `canonical_status` | string | current request status |
| `continuation_path` | string | current continuation classification |
| `resume_required` | bool | whether run continuation is scheduled or already completed |
| `message` | string | optional diagnostic |

### Staff action rules
- Staff-console can submit fallback responses only while request is still open in `waiting|overdue|manual_fallback`.
- Staff-console cannot directly mark `expired`, `resolved`, `continuation_live` or `recovery_resume`.
- Duplicate or stale submissions return typed outcome instead of mutating status.

## Telegram adapter delivery contract
### `TelegramOwnerFeedbackDeliveryEnvelope`
| Field | Type | Required | Notes |
|---|---|---|---|
| `schema_version` | string | yes | `telegram-owner-feedback-v1` |
| `delivery_id` | uuid | yes | dispatch idempotency key |
| `interaction_id` | uuid | yes | canonical request id |
| `interaction_kind` | `decision_request|notify` | yes | `decision_request` for core S17 waits |
| `recipient_provider` | string | yes | always `telegram` in Sprint S17 |
| `recipient_ref` | string | yes | opaque routing ref |
| `content` | `TelegramOwnerFeedbackContent` | yes | rendered question/options |
| `callback_endpoint` | `TelegramOwnerFeedbackCallbackEndpoint` | yes | callback auth + binding handles |
| `wait_window_expires_at` | RFC3339 timestamp | yes | hard owner deadline |
| `manual_fallback_hint` | string | no | short operator-safe text for degraded path |

### `TelegramOwnerFeedbackContent`
| Field | Type | Notes |
|---|---|---|
| `question` | string | required |
| `details_markdown` | string | optional |
| `options` | `TelegramOwnerFeedbackOption[]` | required |
| `allow_free_text` | bool | optional |
| `allow_voice_reply` | bool | optional; normalized to text before callback |
| `free_text_placeholder` | string | optional |

### `TelegramProviderMessageRef`
- Reused unchanged from Sprint S11:
  - opaque provider message identifiers;
  - evidence-only payload;
  - never copied into resume payload or model-visible output.

### `TelegramOwnerFeedbackOption`
| Field | Type | Notes |
|---|---|---|
| `option_id` | string | semantic option id |
| `label` | string | button label |
| `binding_handle` | string | opaque handle resolved to response binding |

### `TelegramOwnerFeedbackCallbackEndpoint`
| Field | Type | Notes |
|---|---|---|
| `url` | string | platform callback endpoint |
| `bearer_token` | string | interaction-scoped callback auth |
| `token_expires_at` | RFC3339 timestamp | request deadline + grace |
| `free_text_binding_handle` | string | optional handle for text/voice reply correlation |

## Telegram callback contract
### Request DTO `TelegramOwnerFeedbackCallbackEnvelope`
| Field | Type | Required | Notes |
|---|---|---|---|
| `schema_version` | string | yes | `telegram-owner-feedback-v1` |
| `interaction_id` | uuid | yes | canonical request id |
| `delivery_id` | uuid | no | original attempt |
| `adapter_event_id` | string | yes | dedupe key |
| `callback_kind` | `delivery_receipt|option_selected|free_text_received|voice_transcript_received|transport_failure` | yes | closed variant |
| `occurred_at` | RFC3339 timestamp | yes | adapter event time |
| `binding_handle` | string | no | required for owner responses |
| `free_text` | string | no | normalized free text; max `8192` UTF-8 bytes |
| `voice_transcript` | string | no | normalized voice transcript; max `8192` UTF-8 bytes |
| `responder_ref` | string | no | opaque actor ref |
| `provider_message_ref` | `TelegramProviderMessageRef` | no | evidence only |
| `provider_update_id` | string | no | evidence only |
| `delivery_status` | `accepted|delivered|failed` | no | for receipts |
| `error` | `TelegramInteractionCallbackError` | no | transport failure detail |

### `TelegramInteractionCallbackError`
| Field | Type | Required | Notes |
|---|---|---|---|
| `code` | string | yes | adapter-typed transport error |
| `retryable` | bool | yes | whether adapter suggests retry |
| `message` | string | no | diagnostic safe for logs |

### Callback response DTO `TelegramOwnerFeedbackCallbackOutcome`
| Field | Type | Notes |
|---|---|---|
| `accepted` | bool | whether request state changed |
| `classification` | `applied|duplicate|stale|expired|invalid` | deterministic outcome |
| `canonical_status` | string | current request status |
| `continuation_path` | `live_same_session|recovery_resume|manual_fallback|expired` | current path |
| `resume_required` | bool | whether live/recovery continuation is scheduled |
| `message` | string | optional diagnostic |

### Telegram callback rules
- `binding_handle` is the only source-of-truth correlator for Telegram responses.
- `voice_transcript_received` is normalized into the same semantic path as free-text, but `response_source_kind=telegram_voice`.
- `duplicate|stale|expired|invalid` return HTTP `200` with typed classification to stop uncontrolled adapter retries.

## Internal gRPC bridge
### `SubmitInteractionCallbackRequest` additions
- `interaction_id`
- `delivery_id`
- `adapter_event_id`
- `callback_kind`
- `binding_handle`
- `free_text`
- `voice_transcript`
- `responder_ref`
- `provider_message_ref_json`
- `provider_update_id`
- `delivery_status`
- `transport_error_code`
- `transport_retryable`
- `raw_payload_json`

### `SubmitInteractionCallbackResponse`
- `accepted`
- `classification`
- `canonical_status`
- `continuation_path`
- `resume_required`
- `effective_response_id`

### `GetRunInteractionResumePayloadResponse`
- `found`
- `payload_json`
- `interaction_id`
- `request_status`
- `continuation_path`
- `response_source_kind`

## Модель ошибок
- Canonical codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- Staff-specific mapping:
  - missing binding or action not allowed -> `failed_precondition`
  - duplicate/stale/expired/invalid submission -> HTTP `200` with typed outcome
- Telegram-specific mapping:
  - invalid callback bearer token -> HTTP `401 unauthorized`
  - malformed schema -> HTTP `400 invalid_argument`
  - unknown request or unknown handle -> HTTP `404 not_found`
  - duplicate/stale/expired/invalid domain outcome -> HTTP `200` with typed classification

## Retries / rate limits
- MCP tool idempotency: `(run_id, mcp_request_id)`
- Staff response idempotency: `(request_id, submission_id)`
- Telegram callback idempotency: `(interaction_id, adapter_event_id)`
- Telegram transport retries remain in `worker`; staff-console write path is synchronous and idempotent, without retry loops in UI.

## Backward compatibility
- S17 is additive relative to Sprint S10/S11 foundation.
- `user.decision.request` input stays compatible; output/resume payload gain additive owner-feedback fields.
- `owner.feedback.request` remains untouched as control tool.
- Existing generic interactions are not retroactively remapped into owner-feedback overlays.

## Наблюдаемость
- Логи:
  - staff request list/detail fetches with `request_id`, `canonical_status`, `projection_state`
  - staff response submissions with `submission_id`, `classification`, `continuation_path`
  - Telegram callback classification with `binding_handle`, `callback_kind`, `classification`
- Метрики:
  - `kodex_owner_feedback_staff_response_total{classification}`
  - `kodex_owner_feedback_callback_total{callback_kind,classification}`
  - `kodex_owner_feedback_response_source_total{response_source_kind}`
- Трейсы:
  - `staff web-console -> api-gateway -> control-plane -> postgres`
  - `telegram-interaction-adapter -> api-gateway -> control-plane -> postgres`

## Открытые вопросы
- OpenAPI/proto/codegen artefacts are intentionally deferred to `run:dev`; this design stage fixes DTO shape and ownership only.

## Апрув
- request_id: owner-2026-03-27-issue-568-api-contract
- Решение: pending
- Комментарий: Ожидается review transport boundaries и typed response submission path.
