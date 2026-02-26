---
doc_id: MON-CK8S-AISLOT-BUILD-0001
type: monitoring
title: "AI Slot Build Pipeline — Monitoring & Observability"
status: active
owner_role: SRE
created_at: 2026-02-26
updated_at: 2026-02-26
related_issues: [205]
related_runbooks: ["RB-CK8S-AISLOT-BUILD-0001"]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "issue-205-ai-repair"
---

# Monitoring & Observability: AI Slot Build Pipeline

## TL;DR
- Основной health signal: доля успешных build/mirror задач для ai-slot.
- Критичный error signal: `MANIFEST_UNKNOWN` и `retrieving image from cache`.
- Primary logs: `codex-k8s-control-plane`, `codex-k8s-worker`.
- Runbook: `docs/ops/runbook_ai_slot_build_failures.md`.

## Источники данных
- Metrics:
  - Kubernetes job status (успех/ошибка).
  - Prometheus `kube_job_status_*` (если включен сбор в кластере).
- Logs:
  - `deploy/codex-k8s-control-plane`;
  - `deploy/codex-k8s-worker`.
- Events:
  - события failed jobs и rollout событий control-plane.

## Оперативные панели (минимум для MVP)
| Название | Источник | Для чего | Owner |
|---|---|---|---|
| Build jobs health | `kubectl get jobs` / Prometheus | Видеть всплеск failed build/mirror jobs | SRE |
| Control-plane error stream | `kubectl logs` / Loki | Отслеживать `MANIFEST_UNKNOWN` и cache-сбои | SRE |
| Worker run recovery | `kubectl logs` / Loki | Проверять run-level impact (повторные падения) | SRE |

## Метрики (каталог)
### Golden signals (pipeline scope)
- Traffic:
  - количество стартовавших build/mirror jobs за 15m.
- Errors:
  - доля failed jobs за 15m/1h;
  - количество `MANIFEST_UNKNOWN` в логах за 15m/1h.
- Latency:
  - p95 времени build job (start -> complete).
- Saturation:
  - backlog build/deploy задач (если доступен runtime queue signal).

### Минимальные kubectl-проверки
```bash
ns="codex-k8s-prod"
kubectl -n "$ns" get jobs --sort-by=.metadata.creationTimestamp | tail -n 30
kubectl -n "$ns" logs deploy/codex-k8s-control-plane --since=30m \
  | grep -E "MANIFEST_UNKNOWN|retrieving image from cache" || true
kubectl -n "$ns" logs deploy/codex-k8s-worker --since=30m \
  | grep -E "MANIFEST_UNKNOWN|build failed|run_id" || true
```

## Логи
- Формат: JSON structured logs.
- Корреляция:
  - `run_id` в worker/control-plane логах;
  - issue id и job name из timeline/issue comments.
- Политика уровней:
  - `INFO`: штатные retries/recovery;
  - `WARN/ERROR`: build/mirror failures, cache errors.

## Проверки и рутины
- На каждом релизе платформы:
  - проверить последние build jobs в ai-slot;
  - убедиться, что нет новых `MANIFEST_UNKNOWN` в окне 30m.
- Ежедневно:
  - проверка failed jobs за сутки;
  - сверка с alert history.

## Связанные документы
- Alerts: `docs/ops/alerts_ai_slot_build_pipeline.md`.
- Rollback: `docs/ops/rollback_plan_ai_slot_build_pipeline.md`.
- SLO: `docs/ops/slo_ai_slot_build_pipeline.md`.
