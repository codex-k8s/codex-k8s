---
doc_id: API-S11-CK8S-0001
type: api-contract
title: "Sprint S11 Day 5 — API contract for Telegram user interaction adapter (Issue #454)"
status: approved
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456, 458]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-454-api-contract"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# API Contract: Sprint S11 Telegram user interaction adapter

## TL;DR
- Контрактный scope: built-in interaction tools from Sprint S10 remain channel-neutral; S11 adds Telegram adapter delivery envelope, normalized callback family, provider message reference contract and continuation semantics.
- Аутентификация: Telegram webhook secret token terminates in adapter contour; adapter -> platform uses interaction-scoped callback bearer token; `api-gateway -> control-plane` uses internal gRPC auth.
- Версионирование: HTTP callback family stays on `/api/v1/...`; adapter delivery/callback payloads versioned as `telegram-interaction-v1`.
- Общий принцип: Telegram-specific payload stays outside agent-facing tool surface; semantic acceptance remains exclusively in `control-plane`.

## Спецификации (source of truth)
- Future OpenAPI source of truth: `services/external/api-gateway/api/server/api.yaml`
- Future gRPC source of truth: `proto/kodex/controlplane/v1/controlplane.proto`
- Design-stage interim source:
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/design_doc.md`
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/api_contract.md`
  - `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`

## Operations / Methods
| Operation | Method/Kind | Path/Name | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Notify user | MCP tool | `user.notify` | run-bound bearer | `(run_id, mcp_request_id)` | channel-neutral, non-blocking |
| Request decision | MCP tool | `user.decision.request` | run-bound bearer | `(run_id, mcp_request_id)` | channel-neutral, enters wait-state |
| Telegram delivery | HTTP POST | adapter-side `/v1/telegram/interaction-deliveries` | adapter credential | `delivery_id` | worker -> adapter contour |
| Telegram callback | HTTP POST | `/api/v1/mcp/interactions/callback` | callback bearer token | `(interaction_id, adapter_event_id)` | adapter -> `api-gateway` |
| Submit callback | gRPC | `SubmitInteractionCallback` | internal service auth | `(interaction_id, adapter_event_id)` | `api-gateway` -> `control-plane` |
| Resume lookup | gRPC | `GetRunInteractionResumePayload` | run-bound bearer | `run_id` | `agent-runner` fetches typed result |

## Agent-facing tool constraints
- Tool names and initial outputs remain those from Sprint S10.
- Additional Telegram-specific validation rules:
  - `response_ttl_seconds` for Telegram path must be within `60..86400`;
  - option labels must remain concise enough for inline keyboard rendering; hard UX target `<= 32` visible chars, validation fallback `<= 64`;
  - if `allow_free_text=true`, the platform must allocate exactly one `free_text_session_handle`.
- Recipient fields stay forbidden in tool input:
  - no `chat_id`
  - no Telegram username
  - no callback handle injection

## Telegram adapter delivery contract (`worker -> adapter`)
### `TelegramInteractionDeliveryEnvelope`
| Field | Type | Required | Notes |
|---|---|---|---|
| `schema_version` | string | yes | `telegram-interaction-v1` |
| `delivery_id` | uuid | yes | dispatch idempotency key |
| `delivery_role` | `primary_dispatch|message_edit|follow_up_notify` | yes | distinguishes initial send from async continuation |
| `interaction_id` | uuid | yes | platform interaction aggregate |
| `interaction_kind` | `notify|decision_request` | yes | |
| `recipient_provider` | string | yes | always `telegram` in Sprint S11 |
| `recipient_ref` | string | yes | opaque platform routing ref, not raw agent input |
| `locale` | string | no | best-effort rendering hint |
| `content` | `TelegramNotifyContent|TelegramDecisionContent` | no | required only for `delivery_role=primary_dispatch` |
| `context_links` | `InteractionContextLinks` | yes | issue/pr/run deep-links |
| `callback_endpoint` | `TelegramCallbackEndpoint` | no | primary dispatch only; required for `decision_request` |
| `provider_message_ref` | `TelegramProviderMessageRef` | no | current provider message context for `message_edit` / follow-up correlation |
| `continuation` | `TelegramInteractionContinuation` | no | required for `delivery_role=message_edit|follow_up_notify` |
| `continuation_policy` | `TelegramContinuationPolicy` | yes | edit vs follow-up rules |
| `delivery_deadline_at` | RFC3339 timestamp | no | primary decision requests only |

### `TelegramCallbackEndpoint`
| Field | Type | Required | Notes |
|---|---|---|---|
| `url` | string (https) | yes | platform callback endpoint |
| `bearer_token` | string | yes | interaction-scoped callback auth |
| `token_expires_at` | RFC3339 timestamp | yes | `response_deadline_at + 24h grace` |
| `handles` | `TelegramCallbackHandle[]` | yes | one per inline option, plus free-text session when enabled |

### `TelegramCallbackHandle`
| Field | Type | Required | Notes |
|---|---|---|---|
| `handle` | string | yes | opaque token, max `48` ASCII chars |
| `handle_kind` | `option|free_text_session` | yes | |
| `button_label` | string | no | required for `option`; adapter uses it for inline keyboard |
| `option_id` | string | no | informational only for rendering/debug, not source-of-truth on callback |
| `expires_at` | RFC3339 timestamp | yes | business deadline |

### `TelegramContinuationPolicy`
| Field | Type | Required | Notes |
|---|---|---|---|
| `preferred_mode` | `edit_in_place_first|follow_up_only` | yes | default `edit_in_place_first` |
| `disable_keyboard_on_resolution` | bool | yes | default `true` |
| `send_follow_up_on_edit_failure` | bool | yes | default `true` |
| `manual_fallback_on_follow_up_failure` | bool | yes | default `true` |

### `TelegramInteractionContinuation`
| Field | Type | Required | Notes |
|---|---|---|---|
| `action` | `edit_message|send_follow_up` | yes | concrete async continuation selected by platform |
| `reason` | `applied_response|edit_failed|expired_wait|operator_fallback` | no | persisted continuation reason |
| `resolution_kind` | `delivery_only|option_selected|free_text_submitted` | no | semantic outcome already accepted by `control-plane` |
| `resolved_at` | RFC3339 timestamp | no | terminal semantic resolution time |

### `TelegramNotifyContent`
| Field | Type | Notes |
|---|---|---|
| `notification_kind` | `completion|next_step|status_update|warning` | mirrors S10 baseline |
| `summary` | string | concise primary text |
| `details_markdown` | string | optional |
| `action_label` | string | optional |
| `action_url` | string | optional https link |

### `TelegramDecisionContent`
| Field | Type | Notes |
|---|---|---|
| `question` | string | required |
| `details_markdown` | string | optional |
| `options` | `TelegramDecisionOption[]` | required |
| `allow_free_text` | bool | optional |
| `free_text_placeholder` | string | optional |
| `expires_at` | RFC3339 timestamp | required |
| `reply_instruction` | string | optional hint for free-text flow |

### `TelegramDecisionOption`
| Field | Type | Notes |
|---|---|---|
| `option_id` | string | stable semantic id |
| `label` | string | user-facing button text |
| `description` | string | optional adapter-side secondary text |
| `callback_handle` | string | opaque handle used in Telegram `callback_data` |

### Adapter immediate ack response
| Field | Type | Notes |
|---|---|---|
| `accepted` | bool | whether adapter accepted delivery request |
| `adapter_delivery_id` | string | adapter-local attempt id |
| `provider_message_ref` | `TelegramProviderMessageRef` | optional provider message identifiers |
| `edit_capability` | `editable|keyboard_only|follow_up_only|unknown` | informs continuation path |
| `retryable` | bool | transport failure may be retried |
| `message` | string | optional diagnostic |

### `TelegramProviderMessageRef`
| Field | Type | Required | Notes |
|---|---|---|---|
| `chat_ref` | string | no | opaque adapter chat ref |
| `message_id` | string | no | provider message id as string |
| `inline_message_id` | string | no | when Telegram uses inline message edit path |
| `sent_at` | RFC3339 timestamp | no | adapter-side send timestamp |

## Telegram callback contract (`adapter -> api-gateway`)
### Request DTO `TelegramInteractionCallbackEnvelope`
| Field | Type | Required | Notes |
|---|---|---|---|
| `schema_version` | string | yes | `telegram-interaction-v1` |
| `interaction_id` | uuid | yes | aggregate id |
| `delivery_id` | uuid | no | link to original attempt |
| `adapter_event_id` | string | yes | stable dedupe key |
| `callback_kind` | `delivery_receipt|option_selected|free_text_received|transport_failure` | yes | closed variant |
| `occurred_at` | RFC3339 timestamp | yes | adapter event time |
| `callback_handle` | string | no | required for `option_selected|free_text_received` |
| `free_text` | string | no | required for `free_text_received`; max `8192` UTF-8 bytes |
| `responder_ref` | string | no | opaque adapter actor ref |
| `provider_message_ref` | `TelegramProviderMessageRef` | no | current provider message context |
| `provider_update_id` | string | no | adapter-carried Telegram update id for evidence |
| `provider_callback_query_id` | string | no | callback query id for evidence |
| `delivery_status` | `accepted|delivered|failed` | no | for `delivery_receipt` |
| `error` | `TelegramInteractionCallbackError` | no | for `transport_failure` |

### `TelegramInteractionCallbackError`
| Field | Type | Required | Notes |
|---|---|---|---|
| `code` | string | yes | adapter-typed transport error |
| `retryable` | bool | yes | whether adapter suggests retry |
| `message` | string | no | diagnostic safe for logs |

### Response DTO `TelegramInteractionCallbackOutcome`
| Field | Type | Notes |
|---|---|---|
| `accepted` | bool | whether callback changed domain state |
| `classification` | `applied|duplicate|stale|expired|invalid` | deterministic outcome |
| `interaction_state` | string | current aggregate state |
| `resume_required` | bool | true when run resume scheduled |
| `continuation_action` | `none|edit_message|send_follow_up|manual_fallback` | async action selected by platform |
| `message` | string | optional diagnostic |

Дополнительные правила callback path:
- `option_id` from adapter is never trusted as source-of-truth; classification resolves only via hashed `callback_handle`.
- If `free_text` exceeds `8192` UTF-8 bytes, callback returns `classification=invalid`, `accepted=false`, without clearing wait-state.
- `duplicate|stale|expired|invalid` return HTTP `200` with typed classification to stop uncontrolled adapter retries.

## Internal gRPC bridge
### `SubmitInteractionCallbackRequest`
- `interaction_id`
- `delivery_id`
- `adapter_event_id`
- `callback_kind`
- `occurred_at`
- `callback_handle`
- `free_text`
- `responder_ref`
- `provider_message_ref_json`
- `provider_update_id`
- `provider_callback_query_id`
- `delivery_status`
- `transport_error_code`
- `transport_retryable`
- `raw_payload_json`

### `SubmitInteractionCallbackResponse`
- `accepted`
- `classification`
- `interaction_state`
- `resume_required`
- `continuation_action`
- `effective_response_id`

### `CompleteInteractionDispatchRequest` additions
- `interaction_id`
- `delivery_attempt_id`
- `adapter_delivery_id`
- `provider_message_ref_json`
- `edit_capability`
- `callback_token_expires_at`

### Continuation dispatch scheduling
- continuation attempts reuse the same worker claim loop as primary delivery;
- `ClaimNextInteractionDispatch` may return a non-primary `delivery_role` inside `request_envelope_json`;
- no отдельный `ScheduleInteractionContinuationRequest` RPC не используется в runtime-контуре этой реализации.

### `GetRunInteractionResumePayloadResponse`
- `found`
- `payload_json`
- `interaction_id`
- `request_status`
- `resolution_reason`

## Error model
- Canonical domain codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- Callback-specific mapping:
  - invalid or expired callback bearer token -> HTTP `401 unauthorized`
  - malformed schema -> HTTP `400 invalid_argument`
  - unknown interaction / unknown delivery binding -> HTTP `404 not_found`
  - duplicate/stale/expired/invalid domain outcome -> HTTP `200` with `accepted=false`
  - persistence or classification failure -> HTTP `500 internal`
- Delivery-specific mapping:
  - adapter immediate rejection with `retryable=false` -> mark attempt failed and raise operator state if critical
  - adapter immediate rejection with `retryable=true` -> worker schedules retry

## Retries / rate limits
- Worker -> adapter:
  - primary delivery retries use attempt ledger from S10 foundation;
  - continuation retries (`edit_message`, `send_follow_up`) use separate `delivery_role` and bounded retry policy.
- Adapter -> platform callback:
  - retry only on transport or HTTP `5xx`;
  - no retry on HTTP `200` with typed classification;
  - callback dedupe key = `(interaction_id, adapter_event_id)`.
- `api-gateway` rate limit:
  - per interaction callback family;
  - should not block legitimate duplicate classification within grace window.

## Backward compatibility / sequencing
- S11 transport additions are additive on top of S10 interaction baseline.
- `run:dev` sequencing:
  1. generic interaction foundation from Sprint S10;
  2. S11 additive Telegram tables and internal RPC fields;
  3. callback HTTP DTO and adapter envelope exposure.
- Approval callback contracts stay separate and are not reused as fallback compatibility layer.

## Наблюдаемость
- Logs include:
  - `interaction_id`
  - `delivery_id`
  - `adapter_event_id`
  - `callback_kind`
  - `classification`
  - `continuation_action`
  - `operator_signal_code`
- Metrics include:
  - callback totals by classification;
  - delivery attempts by role/status;
  - continuation action totals;
  - invalid free-text and handle reuse counts.
- Trace spans:
  - `telegram.delivery.dispatch`
  - `telegram.callback.ingress`
  - `telegram.callback.classification`
  - `telegram.continuation.execute`

## Вопросы, закрытые в `run:plan`
- Notify delivery receipts не вынесены в отдельный execution gate: core wave sequencing оставляет обязательными callback family и transport-failure evidence внутри `S11-E03`/`S11-E05`.
- Отдельный HTTP endpoint для adapter health/bootstrap (`setWebhook`/diagnostics) остаётся вне core platform contract Sprint S11 и не входит в обязательный scope issue `#458`.

## Handover status after `run:plan`
- [x] API contract согласован как baseline для thin-edge `api-gateway` bridge и Telegram adapter contour.
- [x] Waves `S11-E04` и `S11-E05` сохраняют границу `typed transport only` vs `provider-specific webhook/auth`.
- [x] Callback auth lifetime, continuation semantics и typed callback family зафиксированы без Telegram-first drift.
