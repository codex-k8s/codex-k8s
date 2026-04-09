---
doc_id: API-S13-QG-0001
type: api-contract
title: "Quality Governance System — API contract Sprint S13 Day 5"
status: in-review
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [466, 469, 470, 471, 476, 484, 488, 494, 512]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-494-api-contract"
---

# API Contract: Quality Governance System

## TL;DR
- Контрактный scope: internal signals для hidden draft/wave/evidence handoff, staff/private read and decision surfaces, operator gap queue и GitHub service-comment mirror context.
- Аутентификация: run-bound bearer для `agent-runner` callbacks, staff JWT для owner/reviewer/operator read-write surfaces, внутренний service auth для `worker`.
- Версионирование: новые staff/private routes живут в `/api/v1/staff/...`; internal service contracts версионируются в `controlplane.v1`.
- Общий принцип: edge stays thin; only `control-plane` interprets risk/evidence/waiver/release semantics.

## Спецификации (source of truth)
- Future OpenAPI source of truth: `services/external/api-gateway/api/server/api.yaml`
- Future gRPC source of truth: `proto/kodex/controlplane/v1/controlplane.proto`
- Design-stage interim sources:
  - `docs/architecture/initiatives/s13_quality_governance_system/design_doc.md`
  - `docs/architecture/initiatives/s13_quality_governance_system/api_contract.md`

## Operations / Methods
| Operation | Method/Kind | Path/Name | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Report hidden working draft signal | gRPC | `ReportChangeGovernanceDraftSignal` | run-bound bearer | `signal_id` | stores metadata only; raw draft never leaves runner |
| Publish semantic wave map | gRPC | `PublishChangeGovernanceWaveMap` | run-bound bearer | `wave_map_id` | first publishable bridge from hidden draft |
| Upsert evidence block signal | gRPC | `UpsertChangeGovernanceEvidenceSignal` | run-bound bearer | `signal_id` | package or wave scoped evidence input |
| Read change package queue | HTTP GET | `/api/v1/staff/governance/change-packages` | staff JWT | n/a | owner/reviewer/operator list projection |
| Read package detail | HTTP GET | `/api/v1/staff/governance/change-packages/{package_id}` | staff JWT | n/a | waves, evidence, decisions, gaps |
| Submit risk classification decision | HTTP POST | `/api/v1/staff/governance/change-packages/{package_id}/classification` | staff JWT | `decision_id` | explicit risk tier and bundle admissibility |
| Submit waiver decision | HTTP POST | `/api/v1/staff/governance/change-packages/{package_id}/waivers` | staff JWT | `decision_id` | explicit waiver + residual risk |
| Submit release readiness decision | HTTP POST | `/api/v1/staff/governance/change-packages/{package_id}/release-readiness` | staff JWT | `decision_id` | separate from completeness/verification |
| Report governance feedback | HTTP POST | `/api/v1/staff/governance/change-packages/{package_id}/feedback` | staff JWT | `feedback_id` | review/release/postdeploy/remediation gap |
| Read governance gap queue | HTTP GET | `/api/v1/staff/governance/gaps` | staff JWT | n/a | operator backlog and stale gate view |
| Render GitHub status mirror | internal domain op | `UpsertChangeGovernanceStatusComment` | platform contour | `correlation_id` | best-effort read-only mirror |

## Internal callback contract (`agent-runner -> control-plane`)
### `ReportChangeGovernanceDraftSignalRequest`
| Field | Type | Required | Notes |
|---|---|---|---|
| `run_id` | uuid | yes | owning run |
| `signal_id` | string | yes | stable dedupe key |
| `correlation_id` | string | yes | audit correlation |
| `repository_full_name` | string | yes | package namespace |
| `issue_number` | int32 | yes | current issue lineage |
| `pr_number` | int32 | no | current PR lineage if present |
| `branch_name` | string | no | branch carrying hidden draft |
| `draft_ref` | string | yes | internal-only runner draft reference |
| `draft_kind` | `internal_working_draft` | yes | closed enum |
| `change_scope_hints[]` | `ChangeScopeHint[]` | yes | affected bounded contexts and surface kinds |
| `candidate_risk_drivers[]` | `RiskDriver[]` | no | non-authoritative hints |
| `draft_checksum` | string | no | dedupe and replay-safe redaction anchor |
| `occurred_at` | RFC3339 timestamp | yes | UTC |

### `ReportChangeGovernanceDraftSignalResponse`
| Field | Type | Notes |
|---|---|---|
| `package_id` | uuid | canonical change package |
| `draft_state` | `hidden_recorded` | internal-only state acknowledged |
| `next_step_kind` | `wave_map_required|no_op` | package cannot become publishable without wave map |

### `ChangeScopeHint`
| Field | Type | Notes |
|---|---|---|
| `context_key` | string | bounded context or deployable surface |
| `surface_kind` | `domain|transport|schema|release_policy|ui|docs` | where change is expected |

### `RiskDriver`
- Closed enum:
  - `blast_radius`
  - `contract_data`
  - `security_policy`
  - `runtime_release`

### `PublishChangeGovernanceWaveMapRequest`
| Field | Type | Required | Notes |
|---|---|---|---|
| `package_id` | uuid | yes | aggregate root |
| `wave_map_id` | string | yes | idempotency anchor |
| `correlation_id` | string | yes | audit correlation |
| `waves[]` | `SemanticWaveDraft[]` | yes | at least one wave |
| `published_at` | RFC3339 timestamp | yes | UTC |

### `SemanticWaveDraft`
| Field | Type | Notes |
|---|---|---|
| `wave_key` | string | stable package-local key |
| `publish_order` | int32 | monotonic order |
| `dominant_intent` | `code_behavior|schema|transport|ui|ops|mechanical_refactor|docs_only` | closed enum |
| `bounded_scope_kind` | `single_context|cross_context|mechanical_bounded_scope` | used for admissibility |
| `summary` | string | user-safe short description |
| `verification_targets[]` | `VerificationTarget[]` | typed expected checks |

### `VerificationTarget`
| Field | Type | Notes |
|---|---|---|
| `target_kind` | `unit|integration|contract|regression|rollback|release_readiness` | closed enum |
| `target_ref` | string | stable test/check identifier |

### `UpsertChangeGovernanceEvidenceSignalRequest`
| Field | Type | Required | Notes |
|---|---|---|---|
| `package_id` | uuid | yes | aggregate root |
| `signal_id` | string | yes | idempotency anchor |
| `correlation_id` | string | yes | audit correlation |
| `scope_kind` | `package|wave` | yes | evidence attachment scope |
| `scope_ref` | string | yes | package id or wave key |
| `block_kind` | `intent_contract|verification|review_waiver|release_readiness|runtime_feedback` | yes | mandatory evidence family |
| `artifact_links[]` | `GovernanceArtifactLinkInput[]` | yes | issue/PR/run/doc/check references |
| `verification_state_hint` | `not_started|in_progress|met|failed` | no | non-authoritative hint |
| `required_by_tier` | bool | yes | whether the sender believes block mandatory |
| `occurred_at` | RFC3339 timestamp | yes | UTC |

### `GovernanceArtifactLinkInput`
| Field | Type | Notes |
|---|---|---|
| `artifact_kind` | `issue|pull_request|run|agent_session|document|service_comment|release_note` | closed enum |
| `artifact_ref` | string | stable external or internal reference |
| `relation_kind` | `primary_context|evidence_source|decision_followup|feedback_source` | closed enum |
| `display_label` | string | user-safe label |

### `GovernanceArtifactLinkView`
| Field | Type | Notes |
|---|---|---|
| `artifact_kind` | enum | same closed set as input |
| `artifact_ref` | string | stable external or internal reference |
| `relation_kind` | enum | same closed set as input |
| `display_label` | string | user-safe label |

## Staff/private read contract
### `ChangeGovernancePackageListItem`
| Field | Type | Notes |
|---|---|---|
| `package_id` | uuid | aggregate id |
| `repository_full_name` | string | repo namespace |
| `issue_number` | int32 | primary issue |
| `pr_number` | int32, optional | current PR if present |
| `risk_tier` | `low|medium|high|critical` | explicit classification |
| `bundle_admissibility` | `single_wave|mechanical_bounded_scope|requires_decomposition` | publication discipline |
| `publication_state` | `hidden_draft|wave_map_defined|waves_published|review_ready|release_decided|feedback_open` | queue-level phase |
| `evidence_completeness_state` | `not_started|partial|complete|gapped|waived` | separate construct |
| `verification_minimum_state` | `not_started|in_progress|met|failed|waived` | separate construct |
| `waiver_state` | `none|requested|approved|rejected|expired` | decision state |
| `release_readiness_state` | `not_ready|conditionally_ready|ready|blocked|released` | release surface |
| `governance_feedback_state` | `none|open|reclassified|closed` | feedback loop surface |
| `open_gap_count` | int32 | unresolved gaps |
| `updated_at` | RFC3339 timestamp | latest projection update |

### `ChangeGovernancePackageDetail`
| Field | Type | Notes |
|---|---|---|
| `package` | `ChangeGovernancePackageListItem` | summary |
| `waves[]` | `ChangeGovernanceWaveItem[]` | semantic wave lineage |
| `evidence_blocks[]` | `ChangeGovernanceEvidenceBlock[]` | typed completeness surface |
| `active_decisions[]` | `ChangeGovernanceDecisionSummary[]` | classification, waiver, release readiness |
| `feedback_records[]` | `GovernanceFeedbackRecordView[]` | open and resolved gaps |
| `artifact_links[]` | `GovernanceArtifactLinkView[]` | issue/PR/run/doc lineage |
| `comment_mirror_state` | `not_attempted|synced|pending_retry` | GitHub mirror health |

### `ChangeGovernanceWaveItem`
| Field | Type | Notes |
|---|---|---|
| `wave_key` | string | stable key |
| `publish_order` | int32 | sequence |
| `dominant_intent` | enum | primary semantic class |
| `bounded_scope_kind` | enum | single context vs mechanical bundle |
| `publication_state` | `planned|published|reviewed|merged|superseded` | wave-level status |
| `evidence_completeness_state` | enum | wave-specific completeness |
| `verification_minimum_state` | enum | wave-specific verification |

### `ChangeGovernanceEvidenceBlock`
| Field | Type | Notes |
|---|---|---|
| `block_id` | uuid | evidence row |
| `scope_kind` | `package|wave` | attachment scope |
| `scope_ref` | string | package id or wave key |
| `block_kind` | enum | evidence family |
| `state` | `missing|present|verified|waived|stale` | completeness outcome |
| `required_by_tier` | bool | mandatory flag |
| `verification_state` | `not_started|in_progress|met|failed|waived` | separate verification view |
| `artifact_links[]` | `GovernanceArtifactLinkView[]` | typed references |

### `ChangeGovernanceDecisionSummary`
| Field | Type | Notes |
|---|---|---|
| `decision_kind` | `risk_classification|reclassification|waiver|release_readiness` | closed enum |
| `state` | `proposed|approved|rejected|superseded` | decision lifecycle |
| `actor_kind` | `owner|reviewer|operator|system` | who recorded decision |
| `recorded_at` | RFC3339 timestamp | UTC |
| `residual_risk_tier` | `low|medium|high|critical`, optional | only for waiver/release paths |
| `summary` | string | user-safe decision text |

### `GovernanceGapQueueItem`
| Field | Type | Notes |
|---|---|---|
| `package_id` | uuid | owning package |
| `gap_id` | uuid | feedback record id |
| `gap_kind` | `under_classified|missing_evidence|verification_bypass|silent_waiver_attempt|semantic_mix|late_reclassification` | closed enum |
| `severity` | `medium|high|critical` | operator priority |
| `state` | `open|acknowledged|reclassified|closed` | lifecycle |
| `opened_at` | RFC3339 timestamp | UTC |
| `suggested_action` | `reclassify|request_evidence|record_waiver|block_release|close_gap` | typed CTA |

### `GovernanceFeedbackRecordView`
| Field | Type | Notes |
|---|---|---|
| `gap_id` | uuid | feedback record id |
| `gap_kind` | enum | same closed set as queue item |
| `source_kind` | `review|release|postdeploy|remediation|worker_sweep` | gap origin |
| `severity` | `medium|high|critical` | priority |
| `state` | `open|acknowledged|reclassified|closed` | lifecycle |
| `summary_markdown` | string | user-safe description |
| `suggested_action` | enum | next action |

## Staff/private write contract
### `SubmitRiskClassificationDecisionRequest`
| Field | Type | Required | Notes |
|---|---|---|---|
| `decision_id` | string | yes | idempotency key |
| `risk_tier` | `low|medium|high|critical` | yes | explicit tier |
| `bundle_admissibility` | `single_wave|mechanical_bounded_scope|requires_decomposition` | yes | publication constraint |
| `risk_drivers[]` | `blast_radius|contract_data|security_policy|runtime_release` | yes | closed enum |
| `rationale_markdown` | string | yes | typed user-facing note |
| `expected_projection_version` | int64 | yes | stale write guard |

### `SubmitWaiverDecisionRequest`
| Field | Type | Required | Notes |
|---|---|---|---|
| `decision_id` | string | yes | idempotency key |
| `scope_kind` | `package|wave|evidence_block` | yes | where waiver applies |
| `scope_ref` | string | yes | scope id |
| `decision_state` | `approved|rejected` | yes | explicit outcome |
| `residual_risk_tier` | `low|medium|high|critical` | yes | mandatory when approved |
| `follow_up_issue_number` | int32 | no | remediation linkage |
| `reason_markdown` | string | yes | user-safe explanation |
| `expected_projection_version` | int64 | yes | stale write guard |

### `SubmitReleaseReadinessDecisionRequest`
| Field | Type | Required | Notes |
|---|---|---|---|
| `decision_id` | string | yes | idempotency key |
| `decision_state` | `ready|blocked|conditionally_ready|released` | yes | release gate state |
| `open_blocker_refs[]` | string | no | typed artifact refs |
| `residual_risk_tier` | `low|medium|high|critical` | no | only when conditionally ready |
| `summary_markdown` | string | yes | operator/owner summary |
| `expected_projection_version` | int64 | yes | stale write guard |

### `ReportGovernanceFeedbackRequest`
| Field | Type | Required | Notes |
|---|---|---|---|
| `feedback_id` | string | yes | idempotency key |
| `gap_kind` | enum | yes | typed gap |
| `source_kind` | `review|release|postdeploy|remediation|worker_sweep` | yes | origin of feedback |
| `severity` | `medium|high|critical` | yes | priority |
| `summary_markdown` | string | yes | user-safe summary |
| `suggested_action` | enum | yes | typed next action |

## GitHub service-comment render context
- Internal comment model:
  - `headline`
  - `risk_tier`
  - `bundle_admissibility`
  - `evidence_completeness_state`
  - `verification_minimum_state`
  - `waiver_state`
  - `release_readiness_state`
  - `open_gap_badges[]`
  - `staff_deep_link`
- Normative rules:
  - mirror derives only from persisted projection;
  - raw hidden draft metadata never enters comment template;
  - comment text cannot accept or reject waivers; it only deep-links to staff/private action surface.

## Error model
- Canonical domain codes:
  - `invalid_argument`
  - `unauthorized`
  - `forbidden`
  - `not_found`
  - `conflict`
  - `failed_precondition`
  - `internal`
- Normative mapping examples:
  - missing `signal_id` or empty waves list -> `invalid_argument`
  - invalid run token -> `unauthorized`
  - unknown package -> `not_found`
  - stale `expected_projection_version` -> `conflict`
  - hidden draft without wave map trying to move into review-ready decision -> `failed_precondition`
  - persistence/projection refresh failure -> `internal`

## Retries / idempotency
- Draft and evidence callbacks are idempotent by `signal_id`.
- Wave map publication is idempotent by `wave_map_id`.
- Staff decision writes are idempotent by `decision_id` and guarded by `expected_projection_version`.
- Worker feedback writes are idempotent by `feedback_id`.
- GitHub comment mirror retries must never mutate package semantics; only mirror state changes.

## DTO and enum rules
- Closed enums are mandatory for:
  - `risk_tier`
  - `bundle_admissibility`
  - `publication_state`
  - `evidence block kind/state`
  - `decision kind/state`
  - `gap kind`
  - `suggested_action`
- Explicitly forbidden:
  - `map[string]any`, `[]any`, `any` in public or internal transport DTO;
  - free-form gap kinds or bundle admissibility strings;
  - raw draft content, auth headers or prompt transcripts in any DTO.

## Backward compatibility
- Project is pre-production, so coordinated breaking changes are allowed.
- Day5 still preserves staged rollout discipline:
  - no new public webhook family is introduced;
  - staff/private routes are additive;
  - comment mirror is optional and can lag behind canonical state;
  - worker and UI must deploy only after `control-plane` understands new package aggregate and projections.

## Наблюдаемость
- Logs:
  - `quality_governance.draft_signal.accepted`
  - `quality_governance.wave_map.accepted`
  - `quality_governance.decision_command.accepted`
  - `quality_governance.gap_queue.read`
- Metrics:
  - `quality_governance_signal_callback_total{kind,result}`
  - `quality_governance_decision_command_total{decision_kind,result}`
  - `quality_governance_gap_queue_total{state,severity}`
- Traces:
  - `runner-callback -> control-plane.package-update`
  - `staff-http -> control-plane.decision`
  - `worker -> control-plane.feedback`

## Context7 validation
- Новые внешние библиотеки на Day5 не выбирались.
- Contract baseline опирается на существующие source-of-truth платформы: OpenAPI/gRPC/flow_events, без введения нового transport framework.
