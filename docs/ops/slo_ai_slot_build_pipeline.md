---
doc_id: SLO-CK8S-AISLOT-BUILD-0001
type: slo
title: "AI Slot Build Pipeline — SLO Document"
status: active
owner_role: SRE
created_at: 2026-02-26
updated_at: 2026-02-26
related_services: ["control-plane", "worker", "runtime build/mirror jobs"]
related_issues: [205]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "issue-205-ai-repair"
---

# SLO Document: AI Slot Build Pipeline

## TL;DR
- Критичный пользовательский путь: запуск задачи в ai-слоте с успешной сборкой runtime.
- Основные SLI: build success ratio, recovery MTTR.
- Цели:
  - availability SLO: `>=99.0%` успешных ai-slot build за 7 дней;
  - incident recovery SLO: `<=30m` до восстановления после cache-related или build-ref-related сбоя.
- Error budget: `1.0%` на 7 дней.

## Сервис и границы
- Описание:
  - контур сборки и подготовки runtime для ai-slot запусков.
- Клиенты:
  - пользователи, запускающие agent-run в ai-слотах.
- Зависимости:
  - `control-plane`, `worker`, registry/mirror path, Kubernetes jobs, `codegen-check` job.

## Critical User Journeys (CUJ)
1. CUJ-1: пользователь запускает задачу, build-пайплайн завершается успешно, run стартует.
2. CUJ-2: после инцидента build-пайплайн восстанавливается в пределах целевого MTTR.

## SLIs
### SLI-1: Build Availability
- Определение:
  - `successful_build_and_codegen_jobs / total_build_and_codegen_jobs` для ai-slot за окно измерения.
- Источник:
  - Kubernetes jobs status (+ при наличии Prometheus `kube_job_status_*`).
- Период:
  - rolling 7 days.

### SLI-2: Recovery MTTR
- Определение:
  - время от первого зафиксированного cache/build-ref failure до первого успешного build/codegen-check после mitigation.
- Источник:
  - control-plane/worker logs + job timelines.
- Период:
  - per incident + daily aggregate.

## SLO targets
| CUJ/SLI | SLO target | Окно | Исключения | Примечания |
|---|---:|---|---|---|
| CUJ-1 / SLI-1 | `>= 99.0%` | 7d rolling | maintenance windows | контроль через alerts + build/codegen job stats |
| CUJ-2 / SLI-2 | `<= 30m` | per incident | external registry outage (подтвержденный) | контроль через incident timeline |

## Error Budget
- Формула:
  - `budget = 1 - SLO`.
- Политика расхода:
  - при расходе >50% за 3 дня: freeze risky runtime changes без SRE sign-off;
  - при расходе >80%: только stabilisation/hardening задачи до восстановления.

## Alerting (burn rate)
- Alerts:
  - `AI_SLOT_BUILD_FAILURE_BURST`;
  - `AI_SLOT_BUILD_MANIFEST_UNKNOWN_PERSISTENT`;
  - `AI_SLOT_CODEGEN_CHECK_BUILD_REF_INVALID`;
  - `AI_SLOT_BUILD_FAILURE_RATE_ELEVATED`;
  - `AI_SLOT_BUILD_MTTR_SLO_RISK`.
- Runbook link:
  - `docs/ops/runbook_ai_slot_build_failures.md`.

## Review cadence
- Ежедневно: quick review failure-rate и incident queue.
- Еженедельно: полный SLO review с корректировкой порогов алертов и rollback-триггеров.
