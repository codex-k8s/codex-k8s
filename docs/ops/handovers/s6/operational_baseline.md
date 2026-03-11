---
doc_id: OPS-CK8S-S6-OPS-0001
type: runbook
title: "S6 Ops Operational Baseline (Issue #265)"
status: in-review
owner_role: SRE
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [263, 265]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-265-ops-closeout"
---

# S6 Ops Operational Baseline (Issue #265)

## TL;DR
- Операционный хвост Sprint S6 закрыт: postdeploy handover из issue `#263` конвертирован в зафиксированный ops baseline.
- Обновлены runbook-процедуры, monitoring/alerts/SLO baseline и rollback readiness критерии для production.
- На текущем evidence (postdeploy 2026-03-02) признаков деградации SLA не обнаружено; сохранён контроль residual risks.

## Контекст
- Вход: `docs/ops/handovers/s6/postdeploy_ops_handover.md` (issue `#263`) и `docs/ops/production_runbook.md`.
- Stage continuity:
  `#199 -> #201 -> #216 -> #262 -> #263 -> #265`.
- Scope `run:ops`: только markdown-обновления без изменения runtime-конфигурации и кода.

## Закрытие action items из postdeploy handover

| ID | Статус в issue `#265` | Результат |
|---|---|---|
| `OPS-263-01` | closed | Зафиксирован baseline для slow-start диагностики: `startupProbe` инцидентом считается только при устойчивом росте restart count и/или повторяющихся probe-failures после выхода pod в steady-state. |
| `OPS-263-02` | closed | Формализован SLO burn-rate baseline: short/long окна для page/ticket и эскалация по error budget burn. |
| `OPS-263-03` | closed | Каталог алертов синхронизирован с runbook-процедурами и правилами anti-noise (`for`, `keep_firing_for`, user-impact paging). |

## Runbook baseline

### Когда запускать incident-процедуру
- `CrashLoopBackOff`/`ImagePullBackOff` у критичных компонентов (`codex-k8s`, `codex-k8s-control-plane`, `codex-k8s-worker`, `postgres`).
- Устойчивое нарушение SLO (availability/error/critical latency) дольше alert окна.
- Невозможность стабилизировать сервис в течение 15 минут mitigation-действиями.

### Быстрый triage
1. `kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" get pods,deploy,job -o wide`
2. `kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" logs deploy/codex-k8s-control-plane --tail=200`
3. `kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" logs deploy/codex-k8s-worker --tail=200`
4. `kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" logs deploy/codex-k8s --tail=200`
5. `kubectl -n "$CODEXK8S_PRODUCTION_NAMESPACE" get events --sort-by=.lastTimestamp | tail -n 120`

### Критерии стабилизации
- Все критичные deployment в `Ready`, restart count не растёт.
- `healthz/readyz` стабильно возвращают `alive/ok`.
- Ошибки 5xx и latency возвращаются к baseline в пределах 15 минут после mitigation.

## Monitoring baseline

### Golden signals
- Availability: success-rate `2xx/3xx` для API gateway.
- Latency: p95/p99 для `/api/v1/webhooks/github` и staff API.
- Errors: 5xx rate API gateway + gRPC errors `api-gateway -> control-plane`.
- Saturation: CPU/memory pressure и backlog wait-state (`waiting_mcp`, `waiting_owner_review`).

### Операционные пороги

| Сигнал | Ticket | Page |
|---|---|---|
| Availability | < 99.9% (30m) | < 99.5% (10m) |
| API latency p95 | > 1.2s (30m) | > 2.0s (10m) |
| 5xx error rate | > 1% (15m) | > 3% (5m) |
| Worker backlog | рост > 15m без нисходящего тренда | рост > 30m + деградация пользовательских операций |

## Alerting baseline

| Alert | Тип | Правило anti-noise | Runbook action |
|---|---|---|---|
| `CodexK8sApiAvailabilityLow` | page | `for >= 10m`, `keep_firing_for >= 5m` | Incident triage + rollback decision |
| `CodexK8sApiLatencyHigh` | ticket/page | `for >= 5m`, page только при пересечении критического порога | Проверка downstream/нагрузки, mitigation |
| `CodexK8sErrorRateHigh` | page | `for >= 5m`, `keep_firing_for >= 5m` | Ограничение blast radius, rollback при отсутствии стабилизации |
| `CodexK8sWorkerBacklogHigh` | ticket | `for >= 15m` | Анализ очередей/ретраев и saturation |
| `CodexK8sPostgresUnhealthy` | page | `for >= 2m` | DB incident procedure |

## SLO/SLA impact

### SLO baseline
- `SLO-AVAILABILITY`: не ниже `99.5%` на 30-дневном окне.
- `SLO-LATENCY`: p95 критичных API не выше `1.2s` в steady-state.
- `SLO-ERROR`: доля 5xx не выше `1%` в штатном режиме.

### Burn-rate policy
- Ticket: burn-rate > `2x` на окнах `1h + 6h`.
- Page: burn-rate > `6x` на окнах `5m + 1h`.

### Текущий вывод по SLA
- По postdeploy evidence от `2026-03-02` ускоренного расхода error budget не выявлено.
- Изменения stage `run:ops` не ухудшают SLA, так как ограничены документацией и эксплуатационными процедурами.

## Rollback readiness

### Триггеры
- Page alert не стабилизирован в течение 15 минут.
- Повторный рост 5xx/latency после mitigation.
- Деградация DB health или массовые restart-loop у критичных сервисов.

### Порядок отката
1. Freeze rollout и фиксация инцидента в issue.
2. Откат по порядку зависимостей: edge/frontend -> internal, при необходимости миграционный rollback по policy.
3. Post-rollback verification: health, logs, availability/error/latency.

## Residual risks
- Риск ложных probe-срабатываний на холодном старте остаётся под наблюдением.
- Риск деградации под длительной peak-нагрузкой требует отдельного нагрузочного цикла.

## Owner decision (requested)
- `OD-265-01`: Операционный контур Sprint S6 (`release -> postdeploy -> ops`) считать закрытым.
- `OD-265-02`: Следующий этап continuity выполнить через `run:doc-audit` с отдельной issue без trigger-лейбла.

## Context7 references
- Kubernetes probes (`startupProbe` + handoff to liveness/readiness): `/websites/kubernetes_io`.
- Prometheus alerting (`for`, `keep_firing_for`, anti-flapping): `/prometheus/docs`, `/websites/prometheus_io`.

## Связанные документы
- `docs/delivery/epics/s6/epic-s6-day11-ops-operational-closeout.md`
- `docs/ops/handovers/s6/postdeploy_ops_handover.md`
- `docs/ops/production_runbook.md`
- `docs/templates/runbook.md`
- `docs/templates/monitoring.md`
- `docs/templates/alerts.md`
- `docs/templates/slo.md`
- `docs/templates/rollback_plan.md`
