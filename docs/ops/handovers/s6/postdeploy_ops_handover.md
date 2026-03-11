---
doc_id: OPS-CK8S-S6-PD-0001
type: runbook
title: "S6 Postdeploy Ops Handover (Issue #263)"
status: in-review
owner_role: SRE
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [262, 263, 265]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-s6-postdeploy-ops-handover"
---

# S6 Postdeploy Ops Handover (Issue #263)

## TL;DR
- Runtime после release closeout S6 стабилен в текущем контуре postdeploy-проверки.
- Подготовлен единый операционный пакет для `run:ops`: runbook, monitoring, alerts, SLO и rollback.
- Критичных инцидентов в окне postdeploy не выявлено; остаются наблюдаемые риски cold-start и нагрузки.
- Результаты закрытия операционного хвоста зафиксированы в `docs/ops/handovers/s6/operational_baseline.md` (Issue `#265`).

## Runbook (операционная процедура)

### Когда использовать
- После release/postdeploy при проверке стабильности.
- При алертах по `availability`, `latency` или `error-rate`.

### Быстрые проверки
1. `kubectl -n <ns> get pods,deploy,job -o wide`
2. `kubectl -n <ns> logs deploy/codex-k8s-control-plane --tail=200`
3. `kubectl -n <ns> logs deploy/codex-k8s-worker --tail=200`
4. `kubectl -n <ns> logs deploy/codex-k8s --tail=200`
5. `kubectl -n <ns> get events --sort-by=.lastTimestamp | tail -n 80`
6. `kubectl -n <ns> exec deploy/codex-k8s -- wget -qO- http://127.0.0.1:8080/healthz`
7. `kubectl -n <ns> exec deploy/codex-k8s-control-plane -- wget -qO- http://127.0.0.1:8081/health/readyz`

### Критерии OK
- Все критичные deployment в `Ready` (`availableReplicas == desiredReplicas`).
- Нет restart-loop и новых `CrashLoopBackOff`/`ImagePullBackOff`.
- health/readiness endpoint возвращают `alive/ok`.
- Ошибки в логах не демонстрируют возрастающий тренд.

## Monitoring (golden signals)

### Availability
- Доля успешных ответов API gateway (`2xx/3xx`) в 5м/1ч окнах.
- Up/ready состояние `codex-k8s`, `codex-k8s-control-plane`, `codex-k8s-worker`, `postgres`.

### Latency
- p95/p99 latency для критичных API маршрутов (`/api/v1/webhooks/github`, `/api/v1/staff/*`).
- Время startup readiness для control-plane и api-gateway после rollout.

### Errors
- 5xx rate API gateway.
- Ошибки gRPC вызовов `api-gateway -> control-plane`.
- Ошибки worker-процессинга и failed jobs.

### Saturation
- CPU/memory pressure pod'ов control-plane/worker/api-gateway.
- Queue wait states (`waiting_mcp`, `waiting_owner_review`) и backlog.

## Alerts (paging/ticket)

| Alert | Тип | Условие | Окно | Реакция |
|---|---|---|---|---|
| `CodexK8sApiAvailabilityLow` | page | availability ниже целевого SLO | multi-window | немедленная диагностика gateway/control-plane |
| `CodexK8sApiLatencyHigh` | ticket/page | p95 latency выше порога | 5m/30m | проверка деградации downstream и нагрузки |
| `CodexK8sErrorRateHigh` | page | 5xx rate > порога | 5m | mitigation/rollback decision |
| `CodexK8sWorkerBacklogHigh` | ticket | рост backlog wait-state без снижения | 15m | анализ очередей/ретраев worker |
| `CodexK8sPostgresUnhealthy` | page | `pg_isready`/DB health fail | 2m | DB incident procedure |

### Anti-noise policy
- Для алертов использовать `for` для подавления transient spikes.
- Для шумных сигналов применять `keep_firing_for`, чтобы снизить flap в recovery-фазе.
- Paging только по user-impact сигналам (availability/error/critical latency).

## SLO/SLA impact
- По postdeploy evidence impact на SLA не зафиксирован: критичные сервисы стабильны, restart count = 0.
- SLO budget в текущем окне не демонстрирует ускоренного burn-rate.
- Остаточные риски:
  - transient startup probe warnings на холодном старте;
  - потенциальная деградация при длительной нагрузке вне окна текущего наблюдения.

## Rollback plan (операционный baseline)

### Триггеры rollback
- Устойчивое нарушение SLO (availability/latency/error-rate).
- Рост user-visible ошибок после релиза.
- Невозможность стабилизировать систему mitigation-мерой.

### Порядок
1. Зафиксировать инцидент и freeze на rollout.
2. Откатить сервисы на предыдущий стабильный image-tag по порядку зависимостей.
3. Проверить миграционную совместимость перед rollback схемы.
4. Выполнить post-rollback smoke-check (health, logs, ключевые user-flow).

### Верификация после rollback
- `kubectl get deploy,pod,job` без деградации статусов.
- Нормализация 5xx/latency метрик.
- Отсутствие новых критичных алертов.

## Action items для `run:ops`

| ID | Что сделать | Причина | Владелец | Срок |
|---|---|---|---|---|
| OPS-263-01 | Пересмотреть probe-конфигурации для cold-start сценариев и зафиксировать целевые thresholds | Снижение риска ложных startup warning/restart | SRE | следующий stage |
| OPS-263-02 | Формализовать SLO burn-rate алерты (short/long window) для ключевых API путей | Раннее детектирование деградации и контроль error budget | SRE | следующий stage |
| OPS-263-03 | Добавить runbook-ссылки в каталог алертов и проверить эскалационные маршруты | Сокращение MTTR и единый response flow | SRE + EM | следующий stage |

## Closure status (Issue #265)

| ID | Статус | Артефакт |
|---|---|---|
| `OPS-263-01` | closed | `docs/ops/handovers/s6/operational_baseline.md` |
| `OPS-263-02` | closed | `docs/ops/handovers/s6/operational_baseline.md` |
| `OPS-263-03` | closed | `docs/ops/handovers/s6/operational_baseline.md` |

## Context7 references
- Kubernetes probes guidance: `/websites/kubernetes_io`.
- Prometheus alerting guidance (`for`, `keep_firing_for`, user-impact paging): `/prometheus/docs`, `/websites/prometheus_io`.

## Связанные документы
- `docs/delivery/epics/s6/epic-s6-day10-postdeploy-review.md`
- `docs/delivery/epics/s6/epic-s6-day11-ops-operational-closeout.md`
- `docs/ops/handovers/s6/operational_baseline.md`
- `docs/ops/production_runbook.md`
- `docs/templates/runbook.md`
- `docs/templates/monitoring.md`
- `docs/templates/alerts.md`
- `docs/templates/slo.md`
- `docs/templates/rollback_plan.md`
