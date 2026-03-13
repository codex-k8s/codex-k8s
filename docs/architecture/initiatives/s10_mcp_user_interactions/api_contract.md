---
doc_id: API-S10-CK8S-0001
type: api-contract
title: "Sprint S10 Day 5 — API contract deltas for built-in MCP user interactions (Issue #387)"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387, 389]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-387-api-contract"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# API Contract: Sprint S10 built-in MCP user interactions

## TL;DR
- Контрактный scope: built-in MCP tools `user.notify` / `user.decision.request`, worker -> adapter delivery envelope, adapter -> `api-gateway` callback family и internal gRPC bridge в `control-plane`.
- Аутентификация: run-bound MCP bearer token для built-in tools; interaction-scoped callback bearer token с deadline-aware lifetime и post-deadline grace для adapter callbacks; internal gRPC auth между `api-gateway` и `control-plane`.
- Версионирование: callback transport на `/api/v1/...`; adapter delivery envelope versioned как `v1`.
- Общий принцип: edge остаётся thin adapter; interaction semantics, replay classification и wait-state transitions определяются только в `control-plane`.

## Спецификации (source of truth)
- Future OpenAPI source of truth: `services/external/api-gateway/api/server/api.yaml`
- Future gRPC source of truth: `proto/codexk8s/controlplane/v1/controlplane.proto`
- Design-stage interim source:
  - `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`

## Operations / Methods
| Operation | Method/Kind | Path/Name | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Notify user | MCP tool | `user.notify` | run-bound bearer | `(run_id, mcp_request_id)` | non-blocking |
| Request decision | MCP tool | `user.decision.request` | run-bound bearer | `(run_id, mcp_request_id)` | enters wait-state |
| Adapter delivery | HTTP POST | adapter-side `/v1/interaction-deliveries` | adapter credential | `delivery_id` | worker -> adapter |
| Interaction callback | HTTP POST | `/api/v1/mcp/interactions/callback` | callback bearer token | `(interaction_id, adapter_event_id)` | adapter -> `api-gateway` |
| Submit callback | gRPC | `SubmitInteractionCallback` | internal service auth | `(interaction_id, adapter_event_id)` | `api-gateway` -> `control-plane` |

## Built-in MCP tools
### `user.notify`
#### Input DTO
| Field | Type | Required | Notes |
|---|---|---|---|
| `notification_kind` | `completion|next_step|status_update|warning` | yes | used for audit/classification |
| `summary` | string | yes | 1..200 chars |
| `details_markdown` | string | no | optional extended context |
| `action_label` | string | no | required if `action_url` passed |
| `action_url` | string (https) | no | deep-link/follow-up action |

#### Output DTO
| Field | Type | Notes |
|---|---|---|
| `status` | `accepted` | notify is always async after validation |
| `interaction_id` | uuid | aggregate id |
| `delivery_state` | `queued` | worker owns subsequent delivery attempts |
| `message` | string | optional human-readable hint |

#### Validation rules
- `action_url` без `action_label` -> `invalid_argument`.
- Empty `summary` -> `invalid_argument`.
- Recipient fields in tool input отсутствуют by design.

### `user.decision.request`
#### Input DTO
| Field | Type | Required | Notes |
|---|---|---|---|
| `question` | string | yes | 1..500 chars |
| `details_markdown` | string | no | optional explanatory block |
| `options` | `DecisionOption[]` | yes | 2..5 options |
| `allow_free_text` | bool | no | default `false` |
| `free_text_placeholder` | string | no | only when `allow_free_text=true` |
| `response_ttl_seconds` | int32 | yes | valid response window |

#### `DecisionOption`
| Field | Type | Required | Notes |
|---|---|---|---|
| `option_id` | string | yes | stable machine-readable id |
| `label` | string | yes | user-facing option |
| `description` | string | no | short clarification |

#### Initial output DTO
| Field | Type | Notes |
|---|---|---|
| `status` | `pending_user_response` | request accepted and persisted |
| `interaction_id` | uuid | aggregate id |
| `wait_state` | `waiting_mcp` | coarse runtime state |
| `wait_reason` | `interaction_response` | business meaning |
| `expires_at` | RFC3339 timestamp | decision deadline |

#### Resume payload DTO
| Field | Type | Notes |
|---|---|---|
| `interaction_id` | uuid | same aggregate id |
| `tool_name` | string | always `user.decision.request` |
| `request_status` | `answered|expired|delivery_exhausted|cancelled` | terminal outcome |
| `response_kind` | `option|free_text|none` | `none` for timeout/exhausted/cancelled |
| `selected_option_id` | string | set only for option response |
| `free_text` | string | set only for free-text response |
| `resolved_at` | RFC3339 timestamp | terminal timestamp |
| `resolution_reason` | `accepted|expired|delivery_exhausted|cancelled` | audit-safe final classification |

#### Validation rules
- `options < 2` or `options > 5` -> `invalid_argument`.
- Duplicate `option_id` -> `invalid_argument`.
- `free_text_placeholder` without `allow_free_text=true` -> `invalid_argument`.
- `response_ttl_seconds <= 0` -> `invalid_argument`.
- Existing open interaction wait for the same run -> `failed_precondition`.

## Adapter delivery contract (`worker -> adapter`)
### `InteractionDeliveryEnvelope`
| Field | Type | Required | Notes |
|---|---|---|---|
| `schema_version` | string | yes | `v1` |
| `delivery_id` | uuid | yes | idempotency key for dispatch |
| `interaction_id` | uuid | yes | aggregate id |
| `interaction_kind` | `notify|decision_request` | yes | |
| `recipient_provider` | string | yes | adapter routing key |
| `recipient_ref` | string | yes | opaque adapter destination |
| `locale` | string | no | best-effort locale hint |
| `content` | `NotifyContent|DecisionContent` | yes | discriminated by `interaction_kind` |
| `context_links` | `InteractionContextLinks` | yes | issue/pr/run deep-links |
| `callback_url` | string | yes | platform callback endpoint |
| `callback_bearer_token` | string | yes | interaction-scoped callback auth with post-deadline grace |
| `expires_at` | RFC3339 timestamp | no | required for decision request |

### `NotifyContent`
| Field | Type | Notes |
|---|---|---|
| `notification_kind` | enum | mirrors tool input |
| `summary` | string | concise actionable text |
| `details_markdown` | string | optional |
| `action_label` | string | optional |
| `action_url` | string | optional |

### `DecisionContent`
| Field | Type | Notes |
|---|---|---|
| `question` | string | required |
| `details_markdown` | string | optional |
| `options` | `DecisionOption[]` | required |
| `allow_free_text` | bool | optional |
| `free_text_placeholder` | string | optional |
| `expires_at` | RFC3339 timestamp | required |

### Adapter immediate ack response
| Field | Type | Notes |
|---|---|---|
| `accepted` | bool | whether adapter accepted delivery request |
| `adapter_delivery_id` | string | provider-side message identifier |
| `retryable` | bool | whether immediate failure can be retried |
| `message` | string | optional diagnostic |

## Interaction callback contract (`adapter -> api-gateway`)
### Request DTO `InteractionCallbackEnvelope`
| Field | Type | Required | Notes |
|---|---|---|---|
| `schema_version` | string | yes | `v1` |
| `interaction_id` | uuid | yes | aggregate id |
| `delivery_id` | uuid | no | ties callback to attempt |
| `adapter_event_id` | string | yes | stable callback id for dedupe |
| `callback_kind` | `delivery_receipt|decision_response` | yes | |
| `occurred_at` | RFC3339 timestamp | yes | adapter event time |
| `adapter_delivery_id` | string | no | provider message id |
| `delivery_status` | `accepted|delivered|failed` | no | used for receipt callbacks |
| `response` | `InteractionResponsePayload` | no | required for `decision_response` |
| `error` | `InteractionCallbackError` | no | adapter-side failure detail |

### `InteractionResponsePayload`
| Field | Type | Required | Notes |
|---|---|---|---|
| `response_kind` | `option|free_text` | yes | |
| `selected_option_id` | string | no | required for `option` |
| `free_text` | string | no | required for `free_text` |
| `responder_ref` | string | no | opaque adapter user ref |

### Callback response DTO `InteractionCallbackOutcome`
| Field | Type | Notes |
|---|---|---|
| `accepted` | bool | whether callback changed domain state |
| `classification` | `applied|duplicate|stale|expired|invalid` | deterministic outcome |
| `interaction_state` | string | current aggregate state after processing |
| `resume_required` | bool | true when run resume was scheduled |
| `message` | string | optional diagnostic |

## Internal gRPC bridge
### `SubmitInteractionCallbackRequest`
- `run_id`
- `interaction_id`
- `delivery_id`
- `adapter_event_id`
- `callback_kind`
- `occurred_at`
- `delivery_status`
- `response_kind`
- `selected_option_id`
- `free_text`
- `responder_ref`
- `raw_payload_json`

### `SubmitInteractionCallbackResponse`
- `accepted`
- `classification`
- `interaction_state`
- `resume_required`
- `effective_response_id`

### Worker interaction lifecycle responses
- `CompleteInteractionDispatchResponse`
  - `interaction_id`
  - `interaction_state`
  - `resume_required`
  - `run_id`
  - `resume_correlation_id`
- `ExpireNextInteractionResponse`
  - `found`
  - `interaction_id`
  - `interaction_state`
  - `resume_required`
  - `run_id`
  - `resume_correlation_id`
- Contract note:
  - `resume_required=true` means terminal interaction outcome produced a deterministic resume request, and `worker` must perform idempotent enqueue using `run_id` + `resume_correlation_id`.

## Модель ошибок
- Canonical codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- Tool-specific rules:
  - invalid option set / TTL / empty summary -> `invalid_argument`
  - no resolvable recipient or open conflicting wait -> `failed_precondition`
- Callback-specific rules:
  - invalid bearer token -> HTTP `401 unauthorized`
  - malformed schema -> HTTP `400 invalid_argument`
  - unknown interaction or revoked callback token -> HTTP `404 not_found`
  - duplicate/stale/expired/invalid domain outcome -> HTTP `200` with `accepted=false` and typed `classification`
  - unexpected persistence failure -> HTTP `500 internal`

## Retries / rate limits
- MCP tool retries безопасны по `(run_id, mcp_request_id)`.
- Worker retry policy uses attempt-level exponential backoff; retry decision зависит только от adapter ack и transport error class.
- Callback retries безопасны по `(interaction_id, adapter_event_id)`.
- `api-gateway` rate-limit on callback path should be per `interaction_id` to absorb duplicate adapter retries without global throttling.

## Контракты данных (DTO)
- Typed DTO families:
  - `NotifyRequest`
  - `DecisionRequest`
  - `DecisionOption`
  - `InteractionDeliveryEnvelope`
  - `InteractionCallbackEnvelope`
  - `InteractionResumePayload`
- Запрещено:
  - `map[string]any` / `any` в core callback/output contract;
  - Telegram-specific required fields (`chat_id`, `inline_keyboard`, etc.) в core DTO;
  - free-text duplication в `flow_events` payload.

## Backward compatibility
- Новые tool names additive для built-in MCP catalog.
- Новый callback path additive относительно текущих approval/executor callbacks.
- Внутреннее изменение `wait_reason` требует coordinated rollout:
  - migration backfill `mcp -> approval_pending`;
  - dual-read during cutover;
  - no approval contract reuse.

## Наблюдаемость
- Логи:
  - `interaction.tool.accepted`
  - `interaction.callback.classified`
  - `interaction.callback.rejected`
- Метрики:
  - `interaction_tool_calls_total{tool_name,status}`
  - `interaction_callback_classification_total{classification}`
  - `interaction_delivery_ack_total{adapter,accepted}`
- Трейсы:
  - `mcp tool -> control-plane`
  - `api-gateway callback -> control-plane gRPC`
  - `worker dispatch -> adapter`

## Context7 verification
- Попытка использовать Context7 для `kin-openapi` и `goose` завершилась сообщением `Monthly quota exceeded`.
- Новые зависимости для API contract Day5 не добавляются; используются действующие contract-first conventions репозитория.

## Апрув
- request_id: owner-2026-03-12-issue-387-api-contract
- Решение: pending
- Комментарий: Ожидается review typed contracts перед handover в `run:plan`.
