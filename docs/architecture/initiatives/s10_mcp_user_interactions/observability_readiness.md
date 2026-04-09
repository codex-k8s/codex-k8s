---
doc_id: OBS-S10-CK8S-0001
type: readiness-handover
title: "Sprint S10 Day 7 — Observability and rollout readiness for built-in MCP user interactions (Issue #395)"
status: in-review
owner_role: Dev
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [360, 387, 389, 391, 392, 393, 394, 395]
related_prs: []
related_adrs: ["ADR-0012"]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-395-observability-readiness"
---

# Observability and Rollout Readiness: Sprint S10 built-in MCP user interactions

## TL;DR
- Wave `S10-E05` закрывает обязательный readiness gate перед `run:qa`: в коде добавлены отдельные метрики для request lifecycle, callback classification, dispatch outcomes и resume path.
- Runtime-диагностика теперь опирается на согласованный набор flow events, структурированных логов и correlation keys без утечки `free_text` и channel-specific payloads.
- Rollout discipline остаётся неизменным: `migrations -> control-plane -> worker -> api-gateway`; rollback-first шаги описаны так, чтобы не открывать callback ingress и resume path без подтверждённого evidence.

## Scope

### В коде реализовано
- `control-plane` публикует:
  - `kodex_interaction_requests_created_total{tool_name,interaction_kind}`
  - `kodex_interaction_resume_total{interaction_kind,request_status}`
  - `kodex_interaction_decision_turnaround_seconds{request_status}`
  - persisted snapshot collector: `kodex_interaction_requests_state`, `kodex_interaction_pending_dispatch_backlog`, `kodex_interaction_wait_deadline_overdue`, `kodex_interaction_callback_events_total`, `kodex_interaction_dispatch_attempts_total`
- `worker` публикует:
  - `kodex_interaction_dispatch_attempt_total{adapter,status}`
  - `kodex_interaction_dispatch_retry_scheduled_total{adapter,error_code}`
- `api-gateway` публикует:
  - `kodex_interaction_callback_requests_total{callback_kind,classification}`
  - `kodex_interaction_callback_duration_seconds{callback_kind,classification}`
- Для runtime triage добавлены structured logs по dispatch completion, expiry processing и callback classification.

### Вне scope
- Platform-wide OTel bootstrap/exporter rollout.
- Channel-specific dashboards/adapters.
- Автоматическое создание PrometheusRule/Grafana dashboards в Kubernetes-манифестах текущей волны.

## Метрики и их смысл

| Компонент | Метрика | Что показывает | Как использовать |
|---|---|---|---|
| `control-plane` | `kodex_interaction_requests_created_total{tool_name,interaction_kind}` | сколько interaction requests принято built-in tools | rollout smoke и acceptance evidence по `notify`/`decision_request` |
| `control-plane` | `kodex_interaction_resume_total{interaction_kind,request_status}` | сколько terminal interaction outcomes реально создали resume request | контроль `answered/expired/delivery_exhausted` resume path |
| `control-plane` | `kodex_interaction_decision_turnaround_seconds{request_status}` | latency между созданием decision interaction и terminal outcome | SLI для expiry/backlog regression |
| `control-plane` | `kodex_interaction_requests_state{interaction_kind,state}` | текущее распределение aggregate по состояниям | smoke для open/resolved/expired drift и backlog triage |
| `control-plane` | `kodex_interaction_pending_dispatch_backlog{interaction_kind,queue_kind}` | текущий pending/retry backlog dispatch | контроль queue saturation и retry spillover |
| `control-plane` | `kodex_interaction_callback_events_total{callback_kind,classification}` | persisted callback classifications `applied/duplicate/stale/expired/invalid` | source-of-truth для duplicate/stale alerting |
| `control-plane` | `kodex_interaction_dispatch_attempts_total{interaction_kind,adapter_kind,status}` | persisted dispatch outcomes по adapter/status | контроль exhausted/failed trends независимо от pod restarts |
| `worker` | `kodex_interaction_dispatch_attempt_total{adapter,status}` | распределение delivery attempts по `accepted/failed/exhausted` | проверка delivery outcome и exhaustion spikes |
| `worker` | `kodex_interaction_dispatch_retry_scheduled_total{adapter,error_code}` | накопление retry scheduling по transport errors | early warning для retry backlog и adapter instability |
| `api-gateway` | `kodex_interaction_callback_requests_total{callback_kind,classification}` | callback ingress classifications на edge, включая rejected requests | smoke для callback auth/idempotency edge behavior |
| `api-gateway` | `kodex_interaction_callback_duration_seconds{callback_kind,classification}` | latency callback ingress до typed outcome | smoke после открытия callback path и latency regressions |

## Логи, flow events и trace-correlation

### Flow events
- `interaction.request.created`
- `interaction.dispatch.attempted`
- `interaction.dispatch.retry_scheduled`
- `interaction.callback.received`
- `interaction.response.accepted`
- `interaction.response.rejected`
- `interaction.wait.entered`
- `interaction.wait.resumed`

### Structured logs
- `api-gateway`: `interaction callback handled`
- `worker`: `interaction dispatch completed`
- `worker`: `interaction dispatch failed`
- `worker`: `interaction expiry processed`

### Correlation keys
- `correlation_id` остаётся primary stitch между `control-plane` и `worker` flow events.
- `interaction_id` связывает aggregate, callback, response record и resume lifecycle.
- `delivery_id` и `adapter_event_id` нужны для replay/duplicate triage.
- `run_id` и `resume_correlation_id` фиксируют deterministic resume path.
- `free_text`, `details_markdown` и callback raw payload не логируются в structured logs.

## Runtime diagnostics

### Базовые `/metrics` probes
1. `kubectl -n <candidate-namespace> port-forward deploy/kodex-control-plane 18081:8081`
2. `curl -s http://127.0.0.1:18081/metrics | rg 'kodex_interaction_(requests_created_total|resume_total|decision_turnaround_seconds|requests_state|pending_dispatch_backlog|wait_deadline_overdue|callback_events_total|dispatch_attempts_total)'`
3. `kubectl -n <candidate-namespace> port-forward deploy/kodex-worker 18082:8082`
4. `curl -s http://127.0.0.1:18082/metrics | rg 'kodex_interaction_dispatch_'`
5. `kubectl -n <candidate-namespace> port-forward deploy/kodex 18080:8080`
6. `curl -s http://127.0.0.1:18080/metrics | rg 'kodex_interaction_callback_'`

### Логи
- `kubectl -n <candidate-namespace> logs deploy/kodex-worker --tail=200 | rg 'interaction (dispatch|expiry)'`
- `kubectl -n <candidate-namespace> logs deploy/kodex --tail=200 | rg 'interaction callback handled'`

### SQL / psql evidence
- Retry backlog:

```sql
SELECT
    adapter_kind,
    status,
    COUNT(*) AS attempts
FROM interaction_delivery_attempts
WHERE status IN ('pending', 'failed')
GROUP BY adapter_kind, status
ORDER BY attempts DESC;
```

- Expired/open waits without terminal outcome:

```sql
SELECT
    ir.id,
    ir.run_id,
    ir.state,
    ir.response_deadline_at,
    ar.wait_reason,
    ar.wait_target_ref
FROM interaction_requests ir
JOIN agent_runs ar ON ar.id = ir.run_id
WHERE ir.interaction_kind = 'decision_request'
  AND ir.state IN ('pending_dispatch', 'open')
  AND ir.response_deadline_at < NOW()
ORDER BY ir.response_deadline_at ASC;
```

## Alert suggestions

### Duplicate / stale callback spike
```promql
sum by (callback_kind, classification) (
  rate(kodex_interaction_callback_events_total{classification=~"duplicate|stale|expired"}[15m])
)
```
- Ticket threshold: steady рост выше собственного baseline 15m подряд.
- Page threshold: одновременный рост `duplicate|stale|expired` вместе с падением `applied`.

### Delivery exhaustion
```promql
sum by (adapter) (
  rate(kodex_interaction_dispatch_attempt_total{status="exhausted"}[15m])
)
```
- Ticket threshold: > `0` в течение 15 минут.
- Page threshold: exhaustion сопровождается ростом open interactions или user-facing blocking waits.

### Callback latency regression
```promql
histogram_quantile(
  0.95,
  sum by (le, callback_kind, classification) (
    rate(kodex_interaction_callback_duration_seconds_bucket[15m])
  )
)
```
- Ticket threshold: p95 > 1s 15 минут подряд.
- Page threshold: p95 > 3s вместе с ростом invalid/expired classifications.

### Resume / expiry readiness
```promql
sum by (interaction_kind, request_status) (
  rate(kodex_interaction_resume_total{interaction_kind="decision_request",request_status=~"expired|delivery_exhausted"}[30m])
)
```
- Использовать вместе с SQL-проверкой open interactions, чтобы отлавливать залипание expiry scan или resume scheduling.

## Rollout discipline

### Обязательная последовательность
1. `migrations`
2. `control-plane`
3. `worker`
4. `api-gateway`

### Что проверяем после каждого шага
- После `migrations`: interaction tables и wait-linkage columns существуют; backfill `wait_reason='approval_pending'` завершён без drift.
- После `control-plane`: видны `kodex_interaction_requests_created_total`, `kodex_interaction_resume_total` и collector family `kodex_interaction_requests_state`; новые tools остаются скрыты rollout gate-ом.
- После `worker`: видны dispatch/retry metrics и structured logs по completion/expiry; terminal outcomes schedule-ят resume без duplicate completion.
- После `api-gateway`: callback metrics/latency появляются, а `classification=duplicate|stale|expired|invalid` даёт typed outcome вместо silent drop.

## Rollback / mitigation

### Kill switches
- Не открывать user-facing tools в MCP catalog до green evidence по metrics/logs.
- Не включать callback adapter traffic, если нет callback classification metrics и worker retry diagnostics.

### Порядок rollback
1. Остановить новое interaction traffic exposure.
2. Откатить `api-gateway`, если callback ingress даёт invalid/5xx spike.
3. Откатить `worker`, если exhaustion/retry path уходит в drift.
4. Откатить `control-plane` write path только после подтверждения, что additive schema остаётся совместимой.
5. DDL не удалять: interaction tables/columns остаются и корректируются только forward-fix.

## Acceptance evidence перед `run:qa`
- `notify` happy-path создаёт `kodex_interaction_requests_created_total{tool_name="user.notify",interaction_kind="notify"}` и terminal delivery evidence без retry drift.
- `decision_request` happy-path создаёт `pending_dispatch -> open -> resolved` и `kodex_interaction_resume_total{interaction_kind="decision_request",request_status="answered"}`.
- Duplicate callback отражается в `kodex_interaction_callback_events_total{classification="duplicate"}` без второго effective response.
- Expiry path создаёт terminal outcome и `kodex_interaction_resume_total{interaction_kind="decision_request",request_status="expired"}` без зависшего wait target.
- Delivery exhaustion даёт `status="exhausted"` и deterministic resume без silent loss.

## Связанные документы
- `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`
- `docs/delivery/epics/s10/epic-s10-day6-mcp-user-interactions-plan.md`
- `docs/templates/release_plan.md`
- `docs/templates/rollback_plan.md`
- `docs/templates/release_notes.md`
